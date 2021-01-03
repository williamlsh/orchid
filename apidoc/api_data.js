define({ "api": [
  {
    "type": "post",
    "url": "/user/register",
    "title": "注册",
    "group": "用户部分",
    "name": "注册",
    "version": "0.1.0",
    "description": "<p>注册</p>",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "account",
            "description": "<p>帐号</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "email",
            "description": "<p>邮箱</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "passwd",
            "description": "<p>密码</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "{",
          "content": "     {\n\t\t   \"account\":\"test\",\n\t\t   \"email\":\"test@qq.com\",\n\t\t   \"passwd\":\"1123123\"\n     }",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "HTTP/1.1 200 OK",
          "content": "HTTP/1.1 200 OK\n{\n\t\"success\": 1,\n\t\"data\": {\n\t\t\t     \"account\":\"test\",\n\t\t         \"email\":\"test@qq.com\",\n\t\t\t\t \"passwd\":\"1123123\"\n\t\t}\n}",
          "type": "json"
        }
      ]
    },
    "filename": "api/user.go",
    "groupTitle": "用户部分",
    "error": {
      "examples": [
        {
          "title": "Response (fail):",
          "content": "{\n  \"code\": 0\n  \"Msg\": \"错误内容\"\n}",
          "type": "json"
        }
      ]
    }
  },
  {
    "type": "post",
    "url": "/user/veriCode",
    "title": "注册获取验证码",
    "group": "用户部分",
    "name": "注册获取验证码",
    "version": "0.1.0",
    "description": "<p>注册获取验证码</p>",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "email",
            "description": ""
          }
        ]
      },
      "examples": [
        {
          "title": "/user/veriCode?email=1163388086@qq.com",
          "content": "/user/veriCode?email=1163388086@qq.com",
          "type": "json"
        }
      ]
    },
    "filename": "api/user.go",
    "groupTitle": "用户部分",
    "error": {
      "examples": [
        {
          "title": "Response (fail):",
          "content": "{\n  \"code\": 0\n  \"Msg\": \"错误内容\"\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response (success):",
          "content": "{\n  \"code\":\"1\"\n  \"data\":\"成功数据\"\n}",
          "type": "json"
        }
      ]
    }
  }
] });
