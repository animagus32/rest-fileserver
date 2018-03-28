### Restful 文件服务


#### 文件服务首页，展示根目录  
GET /


#### 简单上传页面  
GET /upload


#### 上传文件接口  
POST /v1/upload

|参数|类型|说明|
|:----- |:-------|:-----|
|file|file|上传文件|
|path|string|上传相对路径|

返回值：
``` javascript
{
    "code": 0,
    "message": "",
    "data": {
    "file_uri": "http://fileserver.alajia.cc/测试目录/02.pdf"
    }
}
```

#### 创建目录
POST /v1/mkdir
  
|参数|类型|说明|
|:----- |:-------|:-----|  
|path|string|相对路径|

#### 根据路径获取文件列表  
GET /v1/file/list?path=


#### 删除文件或目录
DELETE /v1/file
  
|参数|类型|说明|
|:----- |:-------|:-----|  
|path|string|相对路径|
|force|bool|true/false,默认false,当文件夹下有文件时是否删除文件夹下所有文件|
