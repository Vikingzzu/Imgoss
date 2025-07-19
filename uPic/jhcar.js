// jhcar_auth.js - Authorization抓取脚本
const url = $request.url;
const headers = $request.headers;

if (url.includes('device-api.xchanger.cn/remote-control/vehicle/status/')) {
    const authorization = headers['Authorization'];
    
    if (authorization) {
        // 保存到剪贴板
        $pasteboard.write(authorization);
        
        // 发送通知
        $notification.post(
            "汽车Token", 
            "Authorization已抓取", 
            `Token: ${authorization.substring(0, 50)}...`
        );
        
        // 控制台日志
        console.log("🚗 几何汽车API调用");
        console.log("📍 URL:", url);
        console.log("🔑 Authorization:", authorization);
        console.log("⏰ 获取时间:", new Date().toLocaleString('zh-CN'));
    } else {
        console.log("❌ 未找到Authorization头");
        console.log("📋 可用Headers:", Object.keys(headers));
    }
}

// 必须调用$done()释放资源
$done({});
