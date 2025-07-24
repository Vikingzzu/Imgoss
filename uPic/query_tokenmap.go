package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/util"
)

// 常量定义，与原代码保持一致
const (
    TokenMapPrefixKey      = "tmp-"                 // map items' key prefix
    TokenMapHashKey        = "tm-hash"              // hash for the whole token map items
    TMapLastEventIDKey     = "tm-last-event-id"     // last event id processed for token map
    TMapCheckedEndBlockKey = "tm-checked-end-block" // end block processed for token map
)

// TokenMapItem - 与原代码结构一致
type TokenMapItem struct {
    RootChainType string `json:"roorChainType"` // 注意：原代码中是拼写错误的 "roorChainType"
    RootToken     string `json:"rootToken"`
    ChildToken    string `json:"childToken"`
    EventID       uint64 `json:"eventId"`
}

func (tmi *TokenMapItem) validate() bool {
    return tmi.RootChainType != "" &&
        tmi.RootToken != "" &&
        tmi.ChildToken != "" &&
        tmi.EventID > 0
}

// TokenMapQuery - 独立查询器
type TokenMapQuery struct {
    db *leveldb.DB
}

// NewTokenMapQuery - 创建查询器
func NewTokenMapQuery(dbPath string) (*TokenMapQuery, error) {
    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to open leveldb: %v", err)
    }

    return &TokenMapQuery{db: db}, nil
}

// Close - 关闭数据库连接
func (tmq *TokenMapQuery) Close() error {
    return tmq.db.Close()
}

// GetTokenMap - 获取所有 TokenMap 数据（与原代码逻辑一致）
func (tmq *TokenMapQuery) GetTokenMap() (map[string][]*TokenMapItem, error) {
    result := make(map[string][]*TokenMapItem)

    items, err := tmq.loadTokenMap()
    if err != nil {
        return nil, err
    }

    for _, v := range items {
        if _, ok := result[v.RootChainType]; !ok {
            result[v.RootChainType] = make([]*TokenMapItem, 0, 1)
        }
        result[v.RootChainType] = append(result[v.RootChainType], v)
    }

    return result, nil
}

// GetTokenMapByRootType - 根据链类型获取数据
func (tmq *TokenMapQuery) GetTokenMapByRootType(rootChainType string) ([]*TokenMapItem, error) {
    resultMap, err := tmq.GetTokenMap()
    if err != nil {
        return nil, err
    }

    if item, ok := resultMap[rootChainType]; ok {
        return item, nil
    }

    return nil, nil
}

// GetTokenMapByRootToken - 根据根代币地址获取单个数据
func (tmq *TokenMapQuery) GetTokenMapByRootToken(rootChainType, rootToken string) (*TokenMapItem, error) {
    key := fmt.Sprintf("%s%s%s", TokenMapPrefixKey, rootChainType, rootToken)
    
    value, err := tmq.db.Get([]byte(key), nil)
    if err != nil {
        if err == leveldb.ErrNotFound {
            return nil, nil
        }
        return nil, err
    }

    return tmq.newTokenMapItem([]byte(key), value)
}

// loadTokenMap - 加载所有 TokenMap 数据（与原代码逻辑一致）
func (tmq *TokenMapQuery) loadTokenMap() ([]*TokenMapItem, error) {
    errorItems := make([][]byte, 0)
    result := make([]*TokenMapItem, 0)

    iter := tmq.db.NewIterator(util.BytesPrefix([]byte(TokenMapPrefixKey)), nil)
    defer iter.Release()

    for iter.Next() {
        if item, err := tmq.newTokenMapItem(iter.Key(), iter.Value()); err == nil {
            result = append(result, item)
        } else {
            errorItems = append(errorItems, iter.Key())
            fmt.Printf("Warning: Failed to parse item with key %s: %v\n", string(iter.Key()), err)
        }
    }

    // 清理错误的数据项
    if len(errorItems) > 0 {
        fmt.Printf("Found %d invalid items, cleaning up...\n", len(errorItems))
        for _, key := range errorItems {
            if err := tmq.db.Delete(key, nil); err != nil {
                fmt.Printf("Warning: Failed to delete invalid key %s: %v\n", string(key), err)
            }
        }
    }

    return result, nil
}

// newTokenMapItem - 解析单个 TokenMap 项
func (tmq *TokenMapQuery) newTokenMapItem(key, value []byte) (*TokenMapItem, error) {
    if !strings.HasPrefix(string(key), TokenMapPrefixKey) {
        return nil, fmt.Errorf("key format is wrong: %s", string(key))
    }

    item := &TokenMapItem{}
    if err := json.Unmarshal(value, item); err != nil {
        return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
    }

    if !item.validate() {
        return nil, fmt.Errorf("item validation failed")
    }

    return item, nil
}

// GetAllKeys - 获取所有 TokenMap 相关的 keys（调试用）
func (tmq *TokenMapQuery) GetAllKeys() ([]string, error) {
    var keys []string

    iter := tmq.db.NewIterator(util.BytesPrefix([]byte(TokenMapPrefixKey)), nil)
    defer iter.Release()

    for iter.Next() {
        keys = append(keys, string(iter.Key()))
    }

    return keys, nil
}

// GetMetadata - 获取元数据信息
func (tmq *TokenMapQuery) GetMetadata() (map[string]interface{}, error) {
    metadata := make(map[string]interface{})

    // 获取 hash
    if hashBytes, err := tmq.db.Get([]byte(TokenMapHashKey), nil); err == nil {
        metadata["hash"] = fmt.Sprintf("%x", hashBytes)
    }

    // 获取最后事件ID
    if eventIDBytes, err := tmq.db.Get([]byte(TMapLastEventIDKey), nil); err == nil {
        metadata["lastEventID"] = string(eventIDBytes)
    }

    // 获取检查的结束块
    if endBlockBytes, err := tmq.db.Get([]byte(TMapCheckedEndBlockKey), nil); err == nil {
        metadata["checkedEndBlock"] = string(endBlockBytes)
    }

    return metadata, nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run query_tokenmap.go <leveldb_path> [command] [args...]")
        fmt.Println("Commands:")
        fmt.Println("  all                           - 查询所有 TokenMap 数据")
        fmt.Println("  chain <rootChainType>        - 查询指定链类型的数据")
        fmt.Println("  token <rootChainType> <rootToken> - 查询指定代币的数据")
        fmt.Println("  keys                         - 列出所有 keys") 
        fmt.Println("  metadata                     - 查询元数据")
        fmt.Println("  count                        - 统计数据量")
        os.Exit(1)
    }

    dbPath := os.Args[1]
    command := "all"
    if len(os.Args) > 2 {
        command = os.Args[2]
    }

    // 创建查询器
    tmq, err := NewTokenMapQuery(dbPath)
    if err != nil {
        log.Fatalf("Failed to create TokenMapQuery: %v", err)
    }
    defer tmq.Close()

    switch command {
    case "all":
        // 查询所有数据
        tokenMaps, err := tmq.GetTokenMap()
        if err != nil {
            log.Fatalf("Failed to get TokenMap: %v", err)
        }

        fmt.Printf("=== All TokenMap Data ===\n")
        for chainType, items := range tokenMaps {
            fmt.Printf("\nChain Type: %s (%d items)\n", chainType, len(items))
            for _, item := range items {
                fmt.Printf("  EventID: %d, RootToken: %s, ChildToken: %s\n",
                    item.EventID, item.RootToken, item.ChildToken)
            }
        }

    case "chain":
        if len(os.Args) < 4 {
            log.Fatal("Usage: go run query_tokenmap.go <dbpath> chain <rootChainType>")
        }
        rootChainType := os.Args[3]
        
        items, err := tmq.GetTokenMapByRootType(rootChainType)
        if err != nil {
            log.Fatalf("Failed to get TokenMap by root type: %v", err)
        }

        fmt.Printf("=== TokenMap for Chain: %s ===\n", rootChainType)
        if len(items) == 0 {
            fmt.Println("No data found")
        } else {
            for _, item := range items {
                fmt.Printf("EventID: %d, RootToken: %s, ChildToken: %s\n",
                    item.EventID, item.RootToken, item.ChildToken)
            }
        }

    case "token":
        if len(os.Args) < 5 {
            log.Fatal("Usage: go run query_tokenmap.go <dbpath> token <rootChainType> <rootToken>")
        }
        rootChainType := os.Args[3]
        rootToken := os.Args[4]
        
        item, err := tmq.GetTokenMapByRootToken(rootChainType, rootToken)
        if err != nil {
            log.Fatalf("Failed to get TokenMap by root token: %v", err)
        }

        fmt.Printf("=== TokenMap for %s:%s ===\n", rootChainType, rootToken)
        if item == nil {
            fmt.Println("No data found")
        } else {
            fmt.Printf("EventID: %d, RootToken: %s, ChildToken: %s\n",
                item.EventID, item.RootToken, item.ChildToken)
        }

    case "keys":
        keys, err := tmq.GetAllKeys()
        if err != nil {
            log.Fatalf("Failed to get keys: %v", err)
        }

        fmt.Printf("=== All TokenMap Keys ===\n")
        for _, key := range keys {
            fmt.Println(key)
        }
        fmt.Printf("Total: %d keys\n", len(keys))

    case "metadata":
        metadata, err := tmq.GetMetadata()
        if err != nil {
            log.Fatalf("Failed to get metadata: %v", err)
        }

        fmt.Printf("=== Metadata ===\n")
        for key, value := range metadata {
            fmt.Printf("%s: %v\n", key, value)
        }

    case "count":
        tokenMaps, err := tmq.GetTokenMap()
        if err != nil {
            log.Fatalf("Failed to get TokenMap: %v", err)
        }

        totalCount := 0
        fmt.Printf("=== Data Count by Chain ===\n")
        for chainType, items := range tokenMaps {
            count := len(items)
            totalCount += count
            fmt.Printf("%s: %d items\n", chainType, count)
        }
        fmt.Printf("Total: %d items\n", totalCount)

    default:
        fmt.Printf("Unknown command: %s\n", command)
        os.Exit(1)
    }
}
