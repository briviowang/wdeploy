package {{.package_prefix}}.{{.entity.EntityType}};
{{.copyright}}
import {{$.package_prefix}}.BaseEntity;
{{range $key,$val:=.entity.ImportList}}
import {{$.package_prefix}}.{{$val}}.{{getField $key}};
{{end}}

import java.util.ArrayList;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

{{if .entity.Comment}}
/*{{.entity.Comment}}*/
{{end}}
public class {{.entity.Name}} extends BaseEntity{

{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
    public ArrayList<{{getFieldType $val}}> {{getField $key}}=new ArrayList<>();
{{else}}
{{if index $.entity.FieldCommentList $key}}
    /*{{index $.entity.FieldCommentList $key}}*/
{{end}}
    public {{getFieldType $val}} {{getField $key}};
{{end}}
{{end}}
    public static {{.entity.Name}} instance;

    public static {{.entity.Name}} getInstance() {
        if (instance == null) {
            instance = new {{.entity.Name}}();
        }
        return instance;
    }

    public {{.entity.Name}}(){       
    }
{{if .entity.ShortName}}

    public String getShortName(){
        return "{{.entity.ShortName}}";
    }
{{end}}

    public {{.entity.Name}}(JSONObject jsonObject){
        fromJson(jsonObject);
    }

    public {{.entity.Name}}(String data){
        fromJson(data);
    }

    public {{.entity.Name}} fromJson(JSONObject jsonObject){
        if(jsonObject==null)return null;

{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
        try{
            JSONArray {{getField $key}}Array=jsonObject.optJSONArray("{{getFieldKey $key}}");
            if({{getField $key}}Array!=null){
                for(int i = 0;i < {{getField $key}}Array.length();i++){
{{if isString $val}}
                    this.{{getField $key}}.add({{getField $key}}Array.getString(i));
{{else}}
                    JSONObject subItemObject = {{getField $key}}Array.getJSONObject(i);
                    {{getFieldType $val}} subItem = new {{getFieldType $val}}();
                    subItem.fromJson(subItemObject);
                    this.{{getField $key}}.add(subItem);
{{end}}
                }
            }
        }catch(JSONException e){
            e.printStackTrace();
        }
{{else if isString $val}}
        if(jsonObject.optString("{{$key}}")!=null)
            this.{{getField $key}}=jsonObject.optString("{{$key}}");
{{else}}
        this.{{getField $key}} =new {{getFieldType $val}}(jsonObject.optJSONObject("{{$key}}"));
{{end}}
{{end}}
        return this;
    }

    public JSONObject toJson(){
        JSONObject jsonObject = new JSONObject();

{{if len .entity.FieldList }}
        try {
            JSONArray itemJSONArray;
{{range $key,$val:=.entity.FieldList}}
{{if isList $key}}
            itemJSONArray = new JSONArray();

            for(int i =0; i< {{getField $key}}.size(); i++){
                {{getFieldType $val}} itemData ={{getField $key}}.get(i);
{{if isString $val}}
                itemJSONArray.put(itemData);
{{else}}
                JSONObject itemJSONObject = itemData.toJson();
                itemJSONArray.put(itemJSONObject);
{{end}}
            }
            jsonObject.put("{{getFieldKey $key}}", itemJSONArray);
{{else if isString $val}}
            if({{getField $key}}!=null) jsonObject.put("{{getFieldKey $key}}", {{getField $key}});
{{else}}
            if({{getField $key}}!=null)
                jsonObject.put("{{getFieldKey $key}}", {{getField $key}}.toJson());
{{end}}
{{end}}
        } catch (JSONException e) {
            e.printStackTrace();
        }
{{end}}        
        return jsonObject;
    }

    public {{.entity.Name}} update({{.entity.Name}} obj){
{{range $key,$val:=.entity.FieldList}}
        if(obj.{{getField $key}}!=null) this.{{getField $key}} = obj.{{getField $key}};
{{end}}
        return this;
    }
}
