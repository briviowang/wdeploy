{{.copyright}}
#import <Foundation/Foundation.h>
#import "{{.prefix}}ApiBase.h"
{{range $key,$val:=.entity.ImportList}}
#import "{{getField $key}}.h"
{{end}}

{{range $key,$val:=.entity.ImportList}}
@class {{getField $key}};
{{end}}

{{if .entity.Comment}}
/*{{.entity.Comment}}*/
{{end}}
@interface {{.entity.Name}} : {{.prefix}}ApiBase
{{if .entity.ShortName}}

- (NSString *) getShortName;
{{end}}

- (NSString *) toJSON;

-({{.entity.Name}} *)update:({{.entity.Name}} *)o;

{{range $key,$val:=.entity.FieldList}}
{{if isList $key }}
@property(strong, nonatomic) NSMutableArray/*{{getFieldType $val}}*/ *{{getField $key}};
{{else}}{{if index $.entity.FieldCommentList $key}}
/*{{index $.entity.FieldCommentList $key}}*/{{end}}
@property(strong, nonatomic) {{getFieldType $val}} *{{getField $key}};
{{end}}
{{end}}

@end
