'use strict';

{{range $index,$entity:=.entityList}}
export class {{$entity.Name}}{
    {{range $field,$v:=$entity.FieldList}}{{if index $entity.FieldCommentList $field}}
    /*{{index $entity.FieldCommentList $field}}*/{{end}}    
    /*{{getFieldType $v}}*/{{getField $field}};{{end}}

    fromJson(o){
        if(o==null)return;
        {{range $key,$val:=$entity.FieldList}}{{if isList $key}}
            var {{getField $key}}Array=o.{{getFieldKey $key}}
            if({{getField $key}}Array!=null){
                for(var i = 0;i < {{getField $key}}Array.length;i++){
        {{if isString $val}}
                    this.{{getField $key}}.push({{getField $key}}Array[i]);
        {{else}}
                    this.{{getField $key}}.push(new {{getFieldType $val}}().fromJson({{getField $key}}Array[i]));
        {{end}}
                }
            }
        {{else if isString $val}}if(o.{{$key}}!=null) this.{{getField $key}}=o.{{$key}}
        {{else}}this.{{getField $key}} =new {{getFieldType $val}}().fromJson(o.{{$key}});{{end}}{{end}}        
    }
    toJson(){
        return JSON.stringify(this)
    }
}
{{end}}

export default class ApiClient {
    constructor() {
    }
    request(url,data,success,fail) {
        var post_url='http://wuji.51yingyi.com/index.php/api/'+url;
        fetch(post_url, {
            method: 'POST',
            headers: {
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: data,
        }).then(response => {
            return response.json()
        }).then(response => {
            console.log("curl "+post_url)
            if(response.status=="1"){
                if(success){
                    success(response);
                }
            }else{
                if(fail){
                    fail(response);
                }
            }
        })
    }
    {{range $name,$api:=.apiList}}
    do{{$name}}(request,success,fail){
        this.request("{{$api.url}}",request.toJson(),success,fail)
    }
    do{{$name}}(request,success){
        this.request("{{$api.url}}",request.toJson(),success,null)
    }
    {{end}}    
}
