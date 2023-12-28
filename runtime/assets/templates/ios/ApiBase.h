#import <Foundation/Foundation.h>

@interface {{.prefix}}ApiBase : NSObject
- (instancetype)fromJSONString:(NSString *)JSONString;

- (instancetype)fromJSON:(NSDictionary *)JSON;

- (NSMutableArray *)arrayFormat:(id)data;

- (NSString *)stringFormat:(id)data;

- (NSString *)JSONItemFormat:(NSString *)format data:(NSString *)data;
@end
