{{.copyright}}
#import "{{.prefix}}ApiClient.h"

@interface {{.prefix}}ApiClient ()
@property(strong, nonatomic) NSString *requestURL;
@property(strong, nonatomic) beforeRequestBlock beforeRequestCallback;
@property(strong, nonatomic) afterRequestBlock afterRequestCallback;
@property(strong, nonatomic) getTokenBlock getTokenBlock;
@property BOOL isHideProgress;
@property BOOL isDisableAfterRequest;
@property(strong, nonatomic) NSDictionary *extraRequestParams;
@end

@implementation {{.prefix}}ApiClient {
}
- (instancetype)init:(beforeRequestBlock)beforeRequest
        afterRequest:(afterRequestBlock)afterRequest
            getToken:(getTokenBlock)getToken
           getApiUrl:(getApiUrlBlock)getApiUrl {
    _beforeRequestCallback = beforeRequest;
    _afterRequestCallback = afterRequest;
    _requestURL = getApiUrl();
    _getTokenBlock = getToken;
    _isHideProgress = _isDisableAfterRequest = NO;
    return self;
}

- (void)updateRequestParams:(NSDictionary *)params {
    self.extraRequestParams = params;
}

- (instancetype)hideProgress {
    _isHideProgress = YES;
    return self;
}

- (instancetype)disableAfterRequest {
    _isDisableAfterRequest = YES;
    return self;
}
- (NSString *)appVersion {
    NSString *versionStr = [[NSBundle mainBundle] infoDictionary][@"CFBundleShortVersionString"];
    NSString *buildStr = [[NSBundle mainBundle] infoDictionary][@"CFBundleVersion"];
    return [NSString stringWithFormat:@"%@.%@", versionStr, buildStr];
}
- (NSString *)appLang {
    NSArray* languages =[[NSUserDefaults standardUserDefaults] objectForKey:@"AppleLanguages"];
    return languages.count>0?languages[0]:@"zh-Hans";
}
- (NSString *)appBundle {
    return [[NSBundle mainBundle] bundleIdentifier];
}
- (NSString *)appToken {
    NSString* token=[NSString stringWithFormat:@"%@", _getTokenBlock()];
    if ([token isEqual:@"(null)"]) {
        token = @"";
    }
    return token;
}
- (NSString *)urlEncode:(id)object {
    NSString *string = [NSString stringWithFormat:@"%@", object];
    NSMutableString *output = [NSMutableString string];
    const unsigned char *source = (const unsigned char *) [string UTF8String];
    int sourceLen = (int) strlen((const char *) source);
    for (int i = 0; i < sourceLen; ++i) {
        const unsigned char thisChar = source[i];
        if (thisChar == ' ') {
            [output appendString:@"+"];
        }
        else if (thisChar == '.' || thisChar == '-' || thisChar == '_' || thisChar == '~' ||
                (thisChar >= 'a' && thisChar <= 'z') ||
                (thisChar >= 'A' && thisChar <= 'Z') ||
                (thisChar >= '0' && thisChar <= '9')) {
            [output appendFormat:@"%c", thisChar];
        }
        else {
            [output appendFormat:@"%%%02X", thisChar];
        }
    }
    return output;
}

- (NSString *)buildHttpQuery:(NSMutableDictionary *)parameters {
    NSMutableArray *parts = [NSMutableArray array];
    for (id key in parameters) {
        id value = parameters[key];
        NSString *part = [NSString stringWithFormat:@"%@=%@", [self urlEncode:key], [self urlEncode:value]];
        [parts addObject:part];
    }
    return [parts componentsJoinedByString:@"&"];
}

- (void)post:(NSString *)url parameters:(NSMutableDictionary *)parameters success:(successBlock)success failure:(failureBlock)failure {
    parameters[@"token"] = [self appToken];
    
    [self.extraRequestParams enumerateKeysAndObjectsUsingBlock:^(NSString *key, id obj, BOOL *stop) {
        if (obj == nil) {
            return;
        }
        if (![key isEqualToString:@"data"]) {
            parameters[key] = obj;
        }
    }];
    parameters[@"version"] = [self appVersion];
    parameters[@"platform"] = @"ios";
    parameters[@"bundle"] = [self appBundle];
    parameters[@"lang"]=[self appLang];
#ifdef DEBUG
    /**
    * 启用NSLogger步骤：
    * 1、配置pod
    * pod 'NSLogger', '~> 1.7.0'
    * 2、初始化
    *  - (void)setupNSLogger {
    *       LoggerSetViewerHost(NULL, NULL, 0);
    *       LoggerSetOptions(NULL, LOGGER_DEFAULT_OPTIONS | kLoggerOption_LogToConsole);
    *   }
    * 3、定义宏
    * #define NSLog(...)  LogMessageF(__FILE__, __LINE__, __PRETTY_FUNCTION__, @"NSLog", 0, __VA_ARGS__)
    * #define NSLoggerAPI(...)  LogMessageF(__FILE__, __LINE__, __PRETTY_FUNCTION__, @"NSLog", 0, __VA_ARGS__)
    *
    */
    #ifdef NSLoggerAPI
        NSLoggerAPI(@"curl %@ -d 'token=%@&data=%@'", url, parameters[@"token"], params[@"data"]);
    #else
        printf("curl %s -d 'platform=ios&token=%s&data=%s'\n", [url UTF8String], [parameters[@"token"] UTF8String], [parameters[@"data"] UTF8String]);
    #endif
#endif
    _beforeRequestCallback(parameters, url, _isHideProgress);
    NSMutableURLRequest *urlRequest = [[NSMutableURLRequest alloc] initWithURL:[NSURL URLWithString:url]];

    [urlRequest addValue:@"api-ios-sdk" forHTTPHeaderField:@"x-client"];
    [urlRequest setHTTPMethod:@"POST"];

    [urlRequest setHTTPBody:[[self buildHttpQuery:parameters] dataUsingEncoding:NSUTF8StringEncoding]];
    NSURLSessionDataTask *dataTask = [[NSURLSession sharedSession] dataTaskWithRequest:urlRequest completionHandler:^(NSData *data, NSURLResponse *urlResponse, NSError *error) {
        NSHTTPURLResponse *httpResponse = (NSHTTPURLResponse *) urlResponse;
        dispatch_async(dispatch_get_main_queue(), ^{
            if (httpResponse.statusCode == 200) {
                NSError *parseError = nil;
                NSMutableDictionary *responseObject = [NSJSONSerialization JSONObjectWithData:data options:0 error:&parseError];
                if (responseObject == nil) {
                    responseObject = [[NSMutableDictionary alloc] init];
                    responseObject[@"result"] = @"服务器返回格式错误!";
                    responseObject[@"status"] = @"502";
                }
                _isHideProgress = NO;
                self.afterRequestCallback(responseObject, url);
                if ([[NSString stringWithFormat:@"%@", responseObject[@"status"]] isEqual:@"1"]) {
                    if (success != nil) {
                        success(responseObject, url);
                    }
                }
                else {
                    if (failure != nil) {
                        failure(responseObject, url);
                    }
                }
            }
            else {
                NSMutableDictionary *response = [NSMutableDictionary new];
                response[@"status"] = @"502";
                response[@"result"] = error.userInfo[@"NSLocalizedDescription"];
                if (failure != nil) {
                    failure(response, url);
                }
                else {
                    self.afterRequestCallback(response, nil);
                }
            }
        });
    }];
    [dataTask resume];
}

{{range $key,$val:=.apiList}}
-(void)do{{$key}}:({{$.prefix}}{{$key}}Request *)request success:(successBlock)success{
    [self do{{$key}}:request success:success failure:nil];
}

-(void)do{{$key}}:({{$.prefix}}{{$key}}Request *)request success:(successBlock)success failure:(failureBlock)failure{
    NSMutableDictionary *parameters = [[NSMutableDictionary alloc] init];
    parameters[@"data"] = [request toJSON];
    [self post:[NSString stringWithFormat:@"%@{{$val.url}}",_requestURL]
          parameters:parameters
          success:success
          failure:failure];
}
{{end}}
@end