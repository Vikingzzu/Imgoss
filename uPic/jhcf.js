// 存储 token 到远程服务
const storeToken = async (token, apiKey) => {
  try {
    const response = await fetch('https://jy-token.vikingzzu.workers.dev/api/store-token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey
      },
      body: JSON.stringify({ token })
    });
    
    const result = await response.json();
    console.log("📤 上报结果:", JSON.stringify(result, null, 2));
    
    if (result.success) {
      console.log('✅ Token上报成功:', result.message);
      return result.data;
    } else {
      console.error('❌ Token上报失败:', result.message);
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('🚫 上报请求失败:', error.message);
    throw error;
  }
};

// 几何汽车 Token 抓取脚本（添加上报功能）
const url = $request.url;
const headers = $request.headers;

if (url.includes('device-api.xchanger.cn/remote-control/vehicle/status/')) {
  console.log("✅ URL匹配成功");
  
  // 检查所有可能的Authorization字段
  const auth1 = headers['Authorization'];
  const auth2 = headers['authorization'];
  const auth3 = headers['AUTHORIZATION'];
  
  console.log("🔍 Authorization检查:");
  
  const authorization = auth1 || auth2 || auth3;
  
  if (authorization) {
      console.log("🎉 找到Authorization: " + authorization);
      
      // 上报Token到远程服务
      const apiKey = 'a15566'; // 您的API密钥
      
      storeToken(authorization, apiKey)
        .then(result => {
          console.log("📊 上报数据:", result);
          
          // 修改通知以显示上报状态
          $notification.post(
              "几何汽车Token", 
              "Authorization已抓取并上报", 
              `Token长度: ${authorization.length}`
          );
          console.log("🔔 通知已发送");
        })
        .catch(error => {
          console.error("⚠️ 上报失败:", error.message);
          
          // 即使上报失败也发送通知
          $notification.post(
              "几何汽车Token", 
              "Authorization已抓取(上报失败)", 
              `Token长度: ${authorization.length}`
          );
          console.log("🔔 通知已发送(上报失败)");
        });
      
  } else {
      console.log("❌ 未找到任何Authorization字段");
      console.log("📋 可用的Headers键: " + Object.keys(headers));
  }
}

console.log("🏁 脚本执行完成，调用$done()");
// 重要：不修改请求，直接放行
$done({});
