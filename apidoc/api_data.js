define({ "api": [
  {
    "type": "post",
    "url": "/user/register",
    "title": "注册",
    "name": "注册",
    "version": "0.1.0",
    "description": "<p>注册</p>",
    "success": {
      "examples": [
        {
          "title": "HTTP/1.1 200 OK",
          "content": "HTTP/1.1 200 OK\n{\n\t\"success\": true,\n\t\"value\": [{\n\t\t\"objectId\": \"bb0894ff26c949898693f7bf6978c61a\",\n\t}, ...]\n}",
          "type": "json"
        }
      ]
    },
    "filename": "api/user.go",
    "group": "/home/youman/GoWork/src/orchid/api/user.go",
    "groupTitle": "/home/youman/GoWork/src/orchid/api/user.go"
  },
  {
    "type": "post",
    "url": "/user/veriCode",
    "title": "注册获取验证码",
    "name": "获取验证码",
    "version": "0.1.0",
    "description": "<p>获取验证码</p>",
    "filename": "api/user.go",
    "group": "/home/youman/GoWork/src/orchid/api/user.go",
    "groupTitle": "/home/youman/GoWork/src/orchid/api/user.go"
  }
] });
