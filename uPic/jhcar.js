const url = $request.url;
const headers = $request.headers;

if (url.includes('device-api.xchanger.cn/remote-control/vehicle/status/')) {
  console.log("✅ URL匹配成功");
  
  // 检查所有可能的Authorization字段
  const auth1 = headers['Authorization'];
  const auth2 = headers['authorization'];
  const auth3 = headers['AUTHORIZATION'];
  
  console.log("🔍 Authorization检查:");
  console.log("- Authorization:", auth1 || "未找到");
  console.log("- authorization:", auth2 || "未找到");
  console.log("- AUTHORIZATION:", auth3 || "未找到");
  
  const authorization = auth1 || auth2 || auth3;
  
  if (authorization) {
      console.log("🎉 找到Authorization:", authorization);
      
      try {
          $pasteboard.write(authorization);
          console.log("📋 已写入剪贴板");
          
          $notification.post(
              "几何汽车Token", 
              "Authorization已抓取", 
              authorization.substring(0, 30) + "..."
          );
          console.log("🔔 通知已发送");
      } catch (e) {
          console.log("❌ 操作失败:", e.toString());
      }
  } else {
      console.log("❌ 未找到任何Authorization字段");
      console.log("📋 可用的Headers键:", Object.keys(headers));
  }
}

console.log("🏁 脚本执行完成，调用$done()");
// 重要：不修改请求，直接放行
$done({});
