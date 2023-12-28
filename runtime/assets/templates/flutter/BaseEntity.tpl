
{{.copyright}}
import 'dart:convert';

class BaseEntity {
  Map<String, dynamic> json_decode(String data) {
    try {
      return json.decode(data);
    } catch (e) {
      return null;
    }
  }
}
