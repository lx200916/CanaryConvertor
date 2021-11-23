
<p align="center">

<img src="icon.png" style="width:72px;border-radius:20px;"/>

</p>

<div align="center">

# CanaryConvertor
Export HttpCanary Records to Postman Collection.Only Zip Achieve Supported.

将HttpCanary导出的HTTP报文转换到Postman集合文件。

</div>

## 🍱 特性
* 将HttpCanary抓包记录存档(zip)导出为Postman集合存档，便于在Postman中调试。
* 支持HTTP 1.0/1.1/2 报文。（注意：Canary导出的HTTP/2 报文非标准的二进制帧，仅为文本化报文)
* 自动转化Response,支持Gzip.
* 导出集合名称为被抓包App包名。

## 🌰 用法
编译后执行：
```
./HcyConverter PATH/TO/ZIP -o ./
```
