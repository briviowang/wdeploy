package {{.package_prefix}}.dao.entity;\n

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.experimental.Accessors;
import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import com.fasterxml.jackson.annotation.*;\n

@Data
@Accessors(chain = true)
@TableName("{{.entity.TABLE_NAME}}")
@JsonInclude(JsonInclude.Include.NON_NULL)
public class {{camel .entity.TABLE_NAME}}Entity {\n

    private static final long serialVersionUID = 1L;\n
{{range $key,$val:=.entity.COLUMN_LIST}}
{{if $val.COLUMN_COMMENT}}
    @ApiModelProperty(value = "{{$val.COLUMN_COMMENT}}")
{{end}}
{{if eq $val.COLUMN_KEY "PRI"}}
    @TableId(value = "{{$val.COLUMN_NAME}}")
{{end}}
    @TableField(value ="{{$val.COLUMN_NAME}}")
{{if eq $val.DATA_TYPE "datetime"}}
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss",timezone = "GMT+8")
{{end}}
    private {{springFieldType $val.DATA_TYPE}} {{camelVar $val.COLUMN_NAME}};\n
{{end}}
}
