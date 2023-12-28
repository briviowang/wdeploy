#import "{{.prefix}}ApiBase.h"

@implementation {{.prefix}}ApiBase {

}
- (instancetype)fromJSONString:(NSString *)JSONString {
    NSDictionary *dictionary = [NSJSONSerialization JSONObjectWithData:[JSONString dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];
    return [self fromJSON:dictionary];
}

- (instancetype)fromJSON:(NSDictionary *)JSON {
    return self;
}

- (NSMutableArray *)arrayFormat:(id)data {
    NSMutableArray *result = [NSMutableArray new];
    if (![data isEqual:@""] && data != nil) {
        [result addObjectsFromArray:(NSMutableArray *) data];
    }
    return result;
}

- (NSString *)stringFormat:(id)data {
    return data == nil ? nil : [[[NSString alloc] initWithFormat:@"%@", data] stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceAndNewlineCharacterSet]];
}

- (NSString *)JSONItemFormat:(NSString *)format data:(NSString *)data {
    if(data==nil){
        data=@"";
    }
    data = [data stringByReplacingOccurrencesOfString:@"\"" withString:@"\\\""];
    data = [data stringByReplacingOccurrencesOfString:@"\r\n" withString:@"\\n"];
    data = [data stringByReplacingOccurrencesOfString:@"\n" withString:@"\\n"];
    data = [data stringByReplacingOccurrencesOfString:@"\r" withString:@"\\n"];
    return [NSString stringWithFormat:format, data];
}

@end
