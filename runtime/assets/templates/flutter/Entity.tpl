{{.copyright}}
import 'dart:convert';

import "{{$.package_prefix}}/BaseEntity.dart";
{{range $key,$val:=.entity.ImportList}}
import "{{$.package_prefix}}/{{$val}}/{{getField $key}}.dart";
{{end}}

{{if .entity.Comment}}
/*{{.entity.Comment}}*/
{{end}}
class {{.entity.Name}} extends BaseEntity{

{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
    List<{{getFieldType $val}}> {{getField $key}}=new List<{{getFieldType $val}}>();
{{else}}
{{if index $.entity.FieldCommentList $key}}
    /*{{index $.entity.FieldCommentList $key}}*/
{{end}}
    {{getFieldType $val}} {{getField $key}};
{{end}}
{{end}}

    {{.entity.Name}}([dynamic json]){
        fromJson(json);
    }
{{if .entity.ShortName}}

    String getShortName(){
        return "{{.entity.ShortName}}";
    }
{{end}}

    {{.entity.Name}} fromJson(dynamic data){
        Map<String,dynamic> json;
        if(data==null)return null;
        if(data.runtimeType==String){
          json=json_decode(data);
        }else if (data is Map){
          json=data;
        }
        if(json==null){
            return this;
        }
{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
        List {{getField $key}}List=json["{{getFieldKey $key}}"];
        if({{getField $key}}List!=null){
            for(int i = 0;i < {{getField $key}}List.length;i++){
{{if isString $val}}
                this.{{getField $key}}.add({{getField $key}}List[i]);
{{else}}
                {{getFieldType $val}} subItem = new {{getFieldType $val}}({{getField $key}}List[i]);
                this.{{getField $key}}.add(subItem);
{{end}}
            }
        }

{{else if isString $val}}
        if(json["{{$key}}"]!=null) this.{{getField $key}}=json["{{$key}}"];
{{else}}
        this.{{getField $key}} =new {{getFieldType $val}}(json["{{$key}}"]);
{{end}}
{{end}}
        return this;
    }

    Map<String, dynamic> toJson(){
        Map<String, dynamic> result = new Map();

{{if len .entity.FieldList }}
        List list;
{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
        list = new List();

        for(int i =0; i< {{getField $key}}.length; i++){
            {{getFieldType $val}} itemData ={{getField $key}}[i];
{{if isString $val}}
            list.add(itemData);
{{else}}
            list.add(itemData.toJson());
{{end}}
        }
        result["{{getFieldKey $key}}"]=list;
{{else if isString $val}}
        if({{getField $key}}!=null) result["{{getFieldKey $key}}"]= {{getField $key}};
{{else}}
        if({{getField $key}}!=null) result["{{getFieldKey $key}}"]= {{getField $key}}.toJson();
{{end}}
{{end}}
{{end}}        
        return result;
    }

    String toString(){
        return json.encode(toJson());
    }

    {{.entity.Name}} update({{.entity.Name}} obj){
{{range $key,$val:=.entity.FieldList}}
        if(obj.{{getField $key}}!=null) this.{{getField $key}} = obj.{{getField $key}};
{{end}}
        return this;
    }
}
