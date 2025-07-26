// 存储 token 到远程服务
const storeToken = (token, apiKey) => {
  return new Promise((resolve, reject) => {
    const requestConfig = {
      url: 'https://workers.val.run/jy-token.vikingzzu/api/store-token',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey
      },
      body: JSON.stringify({ token }),
      timeout: 10,
      insecure: false,
      'auto-redirect': true
    };

    $httpClient.post(requestConfig, (error, response, data) => {
      if (error) {
        console.error('🚫 网络请求失败:', error);
        reject(new Error(`网络请求失败: ${error}`));
      } else {
        try {
          const result = JSON.parse(data);
          console.log("📤 上报结果:", JSON.stringify(result, null, 2));
          
          if (result.success) {
            console.log('✅ Token上报成功:', result.message);
            resolve(result.data);
          } else {
            console.error('❌ Token上报失败:', result.message);
            reject(new Error(result.message));
          }
        } catch (parseError) {
          console.error('📋 响应解析失败:', parseError);
          reject(new Error('响应解析失败'));
        }
      }
    });
  });
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
      
      const funYs = 'a15566';
      
      storeToken(authorization, funYs)
        .then(result => {
          console.log("📊 上报数据:", result);
          
          // 修改通知以显示上报状态
          $notification.post(
              "几何汽车Token", 
              "Authorization已抓取并上报", 
              `Token长度: ${authorization.length}`
          );
          console.log("🔔 通知已发送");
          
          // 延迟调用 $done() 确保通知发送完成
          setTimeout(() => {
            console.log("🏁 脚本执行完成，调用$done()");
            $done({});
          }, 1000);
        })
        .catch(error => {
          console.error("⚠️ 上报失败:", error.message);
          
          // 即使上报失败也发送通知
          $notification.post(
              "几何汽车Token", 
              "Authorization已抓取(上报失败)", 
              `原因: ${error.message}`
          );
          console.log("🔔 通知已发送(上报失败)");
          
          // 延迟调用 $done() 确保通知发送完成
          setTimeout(() => {
            console.log("🏁 脚本执行完成，调用$done()");
            $done({});
          }, 1000);
        });
  } else {
      console.log("❌ 未找到任何Authorization字段");
      console.log("📋 可用的Headers键: " + Object.keys(headers));
      
      // 如果没有找到 authorization，立即调用 $done()
      console.log("🏁 脚本执行完成，调用$done()");
      $done({});
  }
}
