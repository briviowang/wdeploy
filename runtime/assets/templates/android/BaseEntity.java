package {{.package_prefix}};
{{.copyright}}

import org.json.JSONException;
import org.json.JSONObject;

public class BaseEntity {
    public BaseEntity() {
    }

    public BaseEntity(String data) {
        fromJson(data);
    }

    public BaseEntity fromJson(String data) {
        try {
            return fromJson(new JSONObject(data));
        } catch (JSONException e) {
            e.printStackTrace();
        }
        return new BaseEntity();
    }

    public BaseEntity fromJson(JSONObject jsonObject) {
        return this;
    }

    public JSONObject toJson() {
        return new JSONObject();
    }

    public String toString() {
        return toJson().toString();
    }
}
