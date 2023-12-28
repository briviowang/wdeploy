{{.copyright}}
#import "{{.entity.Name}}.h"

@implementation {{.entity.Name}}
{{if .entity.ShortName}}

- (NSString *) getShortName{
    return @"{{.entity.ShortName}}";
}
{{end}}

- (instancetype)fromJSON:(NSDictionary *) JSON{          
    if([JSON isKindOfClass:[NSString class]]){return self;}
    if(JSON.count==0){
        JSON= [NSDictionary new];
    }     
{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
    NSMutableArray *{{getField $key}}Array = [self arrayFormat:JSON[@"{{getFieldKey $key}}"]];
    self.{{getField $key}}=[NSMutableArray new];
    if ({{getField $key}}Array && {{getField $key}}Array.count > 0) {
{{if isString $val}}
       for (NSString *item in {{getField $key}}Array)  
            [self.{{getField $key}} addObject:item];
{{else}}
        for (NSDictionary *item in {{getField $key}}Array)
            [self.{{getField $key}} addObject:[[[{{getFieldType $val}} alloc]init]fromJSON:item]];
{{end}}
    }
{{else if isString $val}}
{{if eq $.entity.EntityType "request"}}
    self.{{getField $key}}=JSON[@"{{getFieldKey $key}}"];
{{else}}
    self.{{getField $key}}=[self stringFormat:JSON[@"{{getFieldKey $key}}"]];
{{end}}
{{else}}
    self.{{getField $key}}=[[{{getFieldType $val}} new] fromJSON:JSON[@"{{getFieldKey $key}}"]];
{{end}}
{{end}}
    return self;
}

- (NSString *) toJSON{
    NSMutableArray *res = [NSMutableArray new];
{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
    NSMutableArray *{{getField $key}}List = [NSMutableArray new];
{{if isString $val}}
    for (NSString *item in self.{{getField $key}})  
        [{{getField $key}}List addObject:[NSString stringWithFormat:@"\"%@\"",item]];
{{else}}
    for ({{getFieldType $val}} *item in self.{{getField $key}})
        [{{getField $key}}List addObject:[item toJSON]];
{{end}}
    [res addObject:[NSString stringWithFormat:@"\"{{getFieldKey $key}}\":[%@]",[{{getField $key}}List componentsJoinedByString:@","]]];
{{else if isString $val}}
    if(self.{{getField $key}}!=nil)
        [res addObject:[self JSONItemFormat:@"\"{{getFieldKey $key}}\":\"%@\"" data:self.{{getField $key}}]];
{{else}}
    if(self.{{getField $key}}!=nil)
        [res addObject:[NSString stringWithFormat:@"\"{{getFieldKey $key}}\":%@",[self.{{getField $key}} toJSON]]];
{{end}}
{{end}}

    return [NSString stringWithFormat:@"{%@}",[res componentsJoinedByString:@","]];
}

-({{.entity.Name}} *)update:({{.entity.Name}} *)o{
{{range $key,$val:=.entity.FieldList}}
    if(o.{{getField $key}}!=nil) self.{{getField $key}} = o.{{getField $key}};
{{end}}
    return self;
}
@end