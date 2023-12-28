{{.copyright}}

import 'dart:collection';
import 'dart:convert';
import 'dart:io';
import 'package:dio/dio.dart';
{{range $key,$val:=.apiList}}
import "{{$.package_prefix}}/request/{{$.prefix}}{{$key}}Request.dart";
{{end}}
import 'package:flutter/foundation.dart';
import 'package:package_info/package_info.dart';

typedef void ApiClientCallback(var data);
typedef String ApiClientGetTokenCallback();

class ApiClient {
  Dio dio = Dio();
  String apiUrl;
  ApiClientGetTokenCallback _apiClientGetTokenCallback;
  ApiClientCallback _responseCallback;
  String _proxy;
  PackageInfo _packageInfo;
  ApiClient setApiUrl(String url) {
    apiUrl = url;
    return this;
  }

  ApiClient setProxy(String proxy) {
    _proxy = proxy;
    return this;
  }

  ApiClient setGetTokenCallback(ApiClientGetTokenCallback callback) {
    _apiClientGetTokenCallback = callback;
    return this;
  }

  ApiClient setResponseCallback(ApiClientCallback callback) {
    _responseCallback = callback;
    return this;
  }

  int intval(dynamic data) {
    if (data == null) return 0;
    return data.runtimeType == String ? int.parse(data) : data;
  }

  httpPost(String api, String data, ApiClientCallback successListener,
      {ApiClientCallback fail, ApiClientCallback complete}) async {
    if (data == null || !data.startsWith("{")) {
      data = '{}';
    }
    String url;
    if (apiUrl.endsWith("/")) {
      url = apiUrl + api;
    } else {
      url = apiUrl + "/" + api;
    }
    Map<String, String> postData = HashMap();
    var token = _apiClientGetTokenCallback();
    postData["token"] = token;
    postData["data"] = data;
    if (Platform.isAndroid) {
      postData["platform"] = "android";
    } else if (Platform.isIOS) {
      postData["platform"] = "ios";
    }
    if (_packageInfo == null) {
      _packageInfo = await PackageInfo.fromPlatform();
    }

    postData['package'] = _packageInfo.packageName;
    postData['version'] = _packageInfo.version;
    postData['build'] = _packageInfo.buildNumber;

    if (kReleaseMode) {
      postData['compile'] = 'release';
    } else {
      postData['compile'] = 'debug';
    }

    ///构造Headers
    Map<String, String> headers = HashMap();
    headers["user-agent"] = _packageInfo.packageName +
        '/' +
        _packageInfo.version +
        '.' +
        _packageInfo.buildNumber;
    Options option = Options();
    option.headers = headers;
    option.connectTimeout = 15000;
    option.contentType = ContentType.parse("application/x-www-form-urlencoded");

    ///否则网络获取接口数据.
    try {
      print("curl $url -d 'token=$token&data=" + data + "'");
      if (_proxy != null) {
        (dio.httpClientAdapter as DefaultHttpClientAdapter).onHttpClientCreate =
            (client) {
          client.findProxy = (uri) {
            print("PROXY " + _proxy);
            return "PROXY " + _proxy;
          };
          client.badCertificateCallback = (cert, host, port) => true;
        };
      }

      dio
          .post(url, data: FormData.from(postData), options: option)
          .then((response) {
        var responseData = json.decode(response.data);
        if (_responseCallback != null) {
          _responseCallback(responseData);
          if (fail == null) {
            fail = _responseCallback;
          }
        }
        if (intval(responseData['status']) == 1) {
          if (successListener != null) {
            successListener(responseData);
          }
        } else {
          if (fail != null) {
            fail(responseData);
          }
        }
        if (complete != null) {
          complete(responseData);
        }
      });
    } catch (e, stack) {
      print(e);
      print(stack);

      var data = {"status": 0, "result": "网络错误:" + e.message};
      if (fail != null) {
        fail(data);
      }
      if (complete != null) {
        complete(data);
      }
    }
  }

{{range $key,$val:=.apiList}}
{{if $val.comment }}
{{$val.comment}}
{{end}}
    void do{{$key}}({{$.prefix}}{{$key}}Request request,ApiClientCallback successListener,{ApiClientCallback fail,ApiClientCallback complete}) {
        httpPost("{{$val.url}}", request.toString(),successListener,fail:fail,complete:complete);
    }
{{end}}
}
