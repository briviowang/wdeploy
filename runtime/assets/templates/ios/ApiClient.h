{{.copyright}}
#import <Foundation/Foundation.h>
{{range $key,$val:=.apiList}}
#import "{{$.prefix}}{{$key}}Request.h"
#import "{{$.prefix}}{{$key}}Response.h"
{{end}}

{{range $key,$val:=.apiList}}
@class {{$.prefix}}{{$key}}Request;
{{end}}

typedef void (^beforeRequestBlock)(NSMutableDictionary *data,NSString *url,BOOL hideProgress);

typedef void (^afterRequestBlock)(NSMutableDictionary *data,NSString *url);
typedef NSString* (^getTokenBlock)();

typedef NSString* (^getApiUrlBlock)();

typedef void (^successBlock)(NSMutableDictionary *data,NSString *url);

typedef void (^failureBlock)(NSMutableDictionary *data,NSString *url);

@interface {{.prefix}}ApiClient: NSObject
- (instancetype)init:(beforeRequestBlock)beforeRequest
        afterRequest:(afterRequestBlock)afterRequest
            getToken:(getTokenBlock)getToken
           getApiUrl:(getApiUrlBlock)getApiUrl;

/**
 * 增加额外POST参数，其中key不能是token、bundle、platform、data，会被忽略
 * */
- (void)updateRequestParams:(NSDictionary *)params;

- (instancetype)hideProgress;

- (instancetype)disableAfterRequest;

#pragma mark 不带failure声明
{{range $key,$val:=.apiList}}
{{if $val.comment}}
{{$val.comment}}
{{end}}
-(void)do{{$key}}:({{$.prefix}}{{$key}}Request *)request success:(successBlock)success;
{{end}}

#pragma mark 带failure声明
{{range $key,$val:=.apiList}}
-(void)do{{$key}}:({{$.prefix}}{{$key}}Request *)request success:(successBlock)success failure:(failureBlock)failure;
{{end}}

@end
