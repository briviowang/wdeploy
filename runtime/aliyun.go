package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/domain"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/jsonc"
)

type AccessKeyItem struct {
	AccessKey    string   `json:"access_key"`
	AccessSecret string   `json:"access_secret"`
	Hosts        []string `json:"hosts"`
}

// AliyunHelper 阿里云
type AliyunHelper struct {
	AccessKey    string
	AccessSecret string
	Config       ProjectConfig
}

type SiteStaticsConfig struct {
	Title    string `json:"title"`
	Sql      string `json:"sql"`
	TotalSql string `json:"total_sql"`
}

type ProjectConfig struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`

	ServerHost string `json:"server_host"`
	ServerUser string `json:"server_user"`
	ServerPort string `json:"server_port"`

	DbHost   string `json:"db_host"`
	DbUser   string `json:"db_user"`
	DbPass   string `json:"db_pass"`
	DbName   string `json:"db_name"`
	DbPort   string `json:"db_port"`
	DbPrefix string `json:"db_prefix"`

	StaticsSql []SiteStaticsConfig `json:"statics_sql"`
}

// NewAliyunHelper 初始化
func NewAliyunHelper() AliyunHelper {
	helper := AliyunHelper{}
	helper.init()
	return helper
}

func (c *AliyunHelper) init() {
	default_config_file := GetRuntimeConfigDir() + "/conf.d/" + argsParams["name"] + ".json"
	configContent := FileGetContents(default_config_file)

	json.Unmarshal(jsonc.ToJSON([]byte(configContent)), &c.Config)
}

func (c *AliyunHelper) initParams(req *http.Request) map[string]interface{} {
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}
func (c *AliyunHelper) getParamsString(params map[string]interface{}, key string) string {
	result := fmt.Sprintf("%v", params[key])
	if params[key] == nil {
		result = ""
	}
	return result
}
func (c *AliyunHelper) getParamsInt(params map[string]interface{}, key string) requests.Integer {
	result := fmt.Sprintf("%v", params[key])
	if params[key] == nil {
		result = ""
	}
	return requests.NewInteger(StringToInt(result))
}

func (c *AliyunHelper) AjaxResult(w http.ResponseWriter, response interface{}) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", JSONEncode(response))
}
func (c *AliyunHelper) GetResourceList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	client, _ := bssopenapi.NewClientWithAccessKey("cn-hangzhou", c.Config.AccessKey, c.Config.AccessSecret)
	request := bssopenapi.CreateQueryAvailableInstancesRequest()

	request.ProductCode = c.getParamsString(params, "ProductCode")
	pageSize := StringToInt(c.getParamsString(params, "PageSize"))
	if pageSize > 0 {
		request.PageSize = requests.NewInteger(pageSize)
	}

	response, _ := client.QueryAvailableInstances(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetEcsInfo(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)
	RegionId := c.getParamsString(params, "RegionId")
	InstanceIds := c.getParamsString(params, "InstanceIds")

	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeInstancesRequest()
	request.RegionId = RegionId
	request.InstanceIds = InstanceIds
	request.PageSize = requests.NewInteger(100)
	response, _ := client.DescribeInstances(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetEcsMonitorInfo(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeInstanceMonitorDataRequest()
	request.RegionId = RegionId
	request.InstanceId = c.getParamsString(params, "InstanceId")
	request.StartTime = c.getParamsString(params, "StartTime")
	request.EndTime = c.getParamsString(params, "EndTime")
	request.Period = c.getParamsInt(params, "Period")
	response, _ := client.DescribeInstanceMonitorData(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) EcsModifyAttribute(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateModifyInstanceAttributeRequest()
	request.RegionId = RegionId
	request.InstanceId = c.getParamsString(params, "InstanceId")
	request.InstanceName = c.getParamsString(params, "InstanceName")
	response, _ := client.ModifyInstanceAttribute(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetEcsPortList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeSecurityGroupAttributeRequest()
	request.RegionId = RegionId
	request.SecurityGroupId = c.getParamsString(params, "SecurityGroupId")
	response, _ := client.DescribeSecurityGroupAttribute(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetEcsPrice(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeRenewalPriceRequest()
	request.RegionId = RegionId
	request.ResourceId = c.getParamsString(params, "ResourceId")
	request.PriceUnit = c.getParamsString(params, "PriceUnit")

	response, _ := client.DescribeRenewalPrice(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetEcsDiskList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := ecs.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeDisksRequest()
	request.RegionId = RegionId
	request.InstanceId = c.getParamsString(params, "InstanceId")
	response, _ := client.DescribeDisks(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetDomainList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := domain.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := domain.CreateQueryDomainListRequest()
	request.RegionId = RegionId
	request.PageSize = requests.NewInteger(100)
	request.PageNum = requests.NewInteger(1)
	response, _ := client.QueryDomainList(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetDomainResolveList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := alidns.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.RegionId = RegionId
	request.PageSize = requests.NewInteger(500)
	request.DomainName = c.getParamsString(params, "DomainName")
	response, _ := client.DescribeDomainRecords(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeDBInstancesRequest()
	request.RegionId = RegionId
	request.PageSize = requests.NewInteger(100)
	response, _ := client.DescribeDBInstances(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsInfo(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeDBInstanceAttributeRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	response, _ := client.DescribeDBInstanceAttribute(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsParams(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeParametersRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	response, _ := client.DescribeParameters(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsIPWhiteList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeDBInstanceIPArrayListRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	response, _ := client.DescribeDBInstanceIPArrayList(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsMonitorData(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeDBInstancePerformanceRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.Key = c.getParamsString(params, "Key")
	request.StartTime = c.getParamsString(params, "StartTime")
	request.EndTime = c.getParamsString(params, "EndTime")
	response, _ := client.DescribeDBInstancePerformance(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsSlowLogList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeSlowLogsRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.StartTime = c.getParamsString(params, "StartTime")
	request.EndTime = c.getParamsString(params, "EndTime")
	request.PageSize = c.getParamsInt(params, "PageSize")
	request.PageNumber = c.getParamsInt(params, "PageNumber")
	response, _ := client.DescribeSlowLogs(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsResourceUsage(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeResourceUsageRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	response, _ := client.DescribeResourceUsage(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsConnectionList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeDBInstanceNetInfoRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	response, _ := client.DescribeDBInstanceNetInfo(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) ModifyRdsDescription(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateModifyDBInstanceDescriptionRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.DBInstanceDescription = c.getParamsString(params, "DBInstanceDescription")
	response, _ := client.ModifyDBInstanceDescription(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsPrice(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeRenewalPriceRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.UsedTime = c.getParamsInt(params, "UsedTime")
	request.TimeType = c.getParamsString(params, "TimeType")

	response, _ := client.DescribeRenewalPrice(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsAccountList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeAccountsRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.PageNumber = c.getParamsInt(params, "PageNumber")
	request.PageSize = c.getParamsInt(params, "PageSize")
	response, _ := client.DescribeAccounts(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetRdsBackupList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := rds.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := rds.CreateDescribeBackupsRequest()
	request.RegionId = RegionId
	request.DBInstanceId = c.getParamsString(params, "DBInstanceId")
	request.PageNumber = c.getParamsInt(params, "PageNumber")
	request.PageSize = c.getParamsInt(params, "PageSize")
	response, _ := client.DescribeBackups(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetSmsTemplateList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := dysmsapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := dysmsapi.CreateQuerySmsTemplateListRequest()
	request.RegionId = RegionId
	request.PageSize = requests.NewInteger(50)
	response, _ := client.QuerySmsTemplateList(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetSmsSignList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := dysmsapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := dysmsapi.CreateQuerySmsSignListRequest()
	request.RegionId = RegionId
	request.PageSize = requests.NewInteger(50)
	response, _ := client.QuerySmsSignList(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetSmsStatics(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := dysmsapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := dysmsapi.CreateQuerySendStatisticsRequest()
	request.RegionId = RegionId
	request.IsGlobe = requests.NewInteger(1)
	request.StartDate = c.getParamsString(params, "StartDate")
	request.EndDate = c.getParamsString(params, "EndDate")
	request.PageIndex = c.getParamsInt(params, "PageIndex")
	request.PageSize = requests.NewInteger(50)
	response, _ := client.QuerySendStatistics(request)
	c.AjaxResult(w, response)
}

type OssBucket struct {
	Bucket    oss.BucketProperties
	Stat      oss.GetBucketStatResult
	ConfigXml string
}

func (c *AliyunHelper) GetOssBucketList(w http.ResponseWriter, req *http.Request) {

	client, _ := oss.New("oss-cn-hangzhou.aliyuncs.com", c.Config.AccessKey, c.Config.AccessSecret)

	response, _ := client.ListBuckets()
	var result []OssBucket
	for _, bucket := range response.Buckets {
		item := OssBucket{}
		item.Bucket = bucket
		item.ConfigXml, _ = client.GetBucketWebsiteXml(bucket.Name)
		item.Stat, _ = client.GetBucketStat(bucket.Name)
		result = append(result, item)
	}
	c.AjaxResult(w, result)
}
func (c *AliyunHelper) GetAccountBalance(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := bssopenapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := bssopenapi.CreateQueryAccountBalanceRequest()
	request.RegionId = RegionId
	response, _ := client.QueryAccountBalance(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetBillOverview(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := bssopenapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := bssopenapi.CreateQueryBillOverviewRequest()
	request.RegionId = RegionId
	request.BillingCycle = c.getParamsString(params, "BillingCycle")
	response, _ := client.QueryBillOverview(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetOrderList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := bssopenapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := bssopenapi.CreateQueryOrdersRequest()
	request.RegionId = RegionId
	request.CreateTimeStart = c.getParamsString(params, "CreateTimeStart")
	request.CreateTimeEnd = c.getParamsString(params, "CreateTimeEnd")
	request.PageNum = c.getParamsInt(params, "PageNum")
	request.PageSize = c.getParamsInt(params, "PageSize")
	request.PaymentStatus = c.getParamsString(params, "PaymentStatus")

	response, _ := client.QueryOrders(request)
	c.AjaxResult(w, response)
}
func (c *AliyunHelper) GetOrderDetail(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := bssopenapi.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := bssopenapi.CreateGetOrderDetailRequest()
	request.RegionId = RegionId
	request.OrderId = c.getParamsString(params, "OrderId")

	response, _ := client.GetOrderDetail(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetCmsMetricList(w http.ResponseWriter, req *http.Request) {
	params := c.initParams(req)

	RegionId := c.getParamsString(params, "RegionId")
	client, _ := cms.NewClientWithAccessKey(RegionId, c.Config.AccessKey, c.Config.AccessSecret)

	request := cms.CreateDescribeMetricListRequest()
	request.RegionId = RegionId
	request.Namespace = c.getParamsString(params, "Namespace")
	request.MetricName = c.getParamsString(params, "MetricName")
	request.Period = c.getParamsString(params, "Period")
	request.StartTime = c.getParamsString(params, "StartTime")
	request.EndTime = c.getParamsString(params, "EndTime")
	request.Dimensions = c.getParamsString(params, "Dimensions")
	request.NextToken = c.getParamsString(params, "NextToken")
	request.Length = c.getParamsString(params, "Length")
	request.Express = c.getParamsString(params, "Express")

	response, _ := client.DescribeMetricList(request)
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) IsAliyunEnable(w http.ResponseWriter, req *http.Request) {
	response := make(map[string]interface{})
	response["enable"] = 0
	if StringLen(c.Config.AccessKey) > 0 && StringLen(c.Config.AccessSecret) > 0 {
		response["enable"] = 1
	}
	c.AjaxResult(w, response)
}

func (c *AliyunHelper) GetSiteStatics(w http.ResponseWriter, req *http.Request) {
	var configList = c.Config.StaticsSql

	var result []interface{}

	if len(configList) > 0 {
		if msRemoteDb == nil {
			msRemoteDb = getRemoteDb()
		}

		commonSql := ">=date_sub(now(), interval 200 day) GROUP BY date order by date asc"
		for key, val := range configList {
			sql := val.Sql
			if strings.LastIndex(val.Sql, ";") == -1 {
				sql += commonSql
			}

			resultItem := make(map[string]interface{})
			resultItem["id"] = "chart-" + fmt.Sprintf("%d", key)
			resultItem["title"] = val.Title
			if !IsEmpty(val.TotalSql) {
				total := msRemoteDb.Query(val.TotalSql)[0]["value"]
				resultItem["title"] = fmt.Sprintf("%v(总计：%v)", resultItem["title"], total)
			}

			res := msRemoteDb.Query(sql)
			resultItem["data"] = res
			result = append(result, resultItem)
		}

		remoteDbConnection.Process.Kill()
		msRemoteDb.DB.Close()
		msRemoteDb = nil
	}

	c.AjaxResult(w, result)
}

func (c *AliyunHelper) index(w http.ResponseWriter, req *http.Request) {
	tplContent := ""

	templatePath := "assets/aliyun/index.html"
	contentBytes, _ := Asset(templatePath)
	tplContent = string(contentBytes)

	tmpl := template.New("template")
	tmpl.Delims("${{", "}}")
	tmpl.Funcs(template.FuncMap{})

	_, err := tmpl.Parse(tplContent)
	CheckErr(err)

	var buff bytes.Buffer
	pageData := map[string]interface{}{
		"port": msPort,
	}
	err = tmpl.Execute(&buff, pageData)
	CheckErr(err)
	fmt.Fprintf(w, "%s", buff.String())
}

/*
rds DescribeSlowLogRecords
rds DescribeResourceUsage
rds DescribeDBInstanceNetInfo
rds DescribeAccounts
rds DescribeRenewalPrice

bssopenapi QueryBillOverview
#openapi操作日志
aliyun actiontrail LookupEvents --region cn-huhehaote
*/
func StartAliyunClient() {
	aliyun_helper := NewAliyunHelper()
	msPort = GetRandomPort()

	http.HandleFunc("/", aliyun_helper.index)
	http.HandleFunc("/api/is_aliyun_enable", aliyun_helper.IsAliyunEnable)
	http.HandleFunc("/api/get_resource_list", aliyun_helper.GetResourceList)
	http.HandleFunc("/api/get_aliyun_data", aliyun_helper.GetAliyunData)
	// ecs
	http.HandleFunc("/api/get_ecs_info", aliyun_helper.GetEcsInfo)
	http.HandleFunc("/api/get_ecs_monitor_info", aliyun_helper.GetEcsMonitorInfo)
	http.HandleFunc("/api/get_ecs_port_list", aliyun_helper.GetEcsPortList)
	http.HandleFunc("/api/ecs_modify_attribute", aliyun_helper.EcsModifyAttribute)
	http.HandleFunc("/api/get_ecs_price", aliyun_helper.GetEcsPrice)
	http.HandleFunc("/api/get_ecs_disk_list", aliyun_helper.GetEcsDiskList)

	//rds
	http.HandleFunc("/api/get_rds_list", aliyun_helper.GetRdsList)
	http.HandleFunc("/api/get_rds_info", aliyun_helper.GetRdsInfo)
	http.HandleFunc("/api/get_rds_params", aliyun_helper.GetRdsParams)
	http.HandleFunc("/api/get_rds_ip_white_list", aliyun_helper.GetRdsIPWhiteList)
	http.HandleFunc("/api/get_rds_monitor_data", aliyun_helper.GetRdsMonitorData)
	http.HandleFunc("/api/get_rds_slow_log_list", aliyun_helper.GetRdsSlowLogList)
	http.HandleFunc("/api/get_rds_storage_info", aliyun_helper.GetRdsResourceUsage)
	http.HandleFunc("/api/get_rds_connection_list", aliyun_helper.GetRdsConnectionList)
	http.HandleFunc("/api/modify_rds_description", aliyun_helper.ModifyRdsDescription)
	http.HandleFunc("/api/get_rds_price", aliyun_helper.GetRdsPrice)
	http.HandleFunc("/api/get_rds_account_list", aliyun_helper.GetRdsAccountList)
	http.HandleFunc("/api/get_rds_backup_list", aliyun_helper.GetRdsBackupList)

	//oss
	http.HandleFunc("/api/get_oss_bucket_list", aliyun_helper.GetOssBucketList)

	//短信
	http.HandleFunc("/api/get_sms_template_list", aliyun_helper.GetSmsTemplateList)
	http.HandleFunc("/api/get_sms_sign_list", aliyun_helper.GetSmsSignList)
	http.HandleFunc("/api/get_sms_statics", aliyun_helper.GetSmsStatics)

	//域名
	http.HandleFunc("/api/get_domain_list", aliyun_helper.GetDomainList)
	http.HandleFunc("/api/get_domain_resolve_list", aliyun_helper.GetDomainResolveList)

	//费用
	http.HandleFunc("/api/get_account_balance", aliyun_helper.GetAccountBalance)
	http.HandleFunc("/api/get_bill_overview", aliyun_helper.GetBillOverview)
	http.HandleFunc("/api/get_order_list", aliyun_helper.GetOrderList)
	http.HandleFunc("/api/get_order_detail", aliyun_helper.GetOrderDetail)
	//云监控
	http.HandleFunc("/api/get_cms_metric_list", aliyun_helper.GetCmsMetricList)

	//网站数据库统计
	http.HandleFunc("/api/get_site_statics", aliyun_helper.GetSiteStatics)

	url := fmt.Sprintf("http://localhost:%v", msPort)
	OpenBrowser(url)
	InfoLog(url)

	http.ListenAndServe(fmt.Sprintf(":%v", msPort), nil)
}

func (c *WebDeployService) cmdAliyun() {
	cmd := "aliyun"
	c.AddHelp(cmd, `显示阿里云服务器配置信息`)

	if argsAct != cmd {
		return
	}
	StartAliyunClient()
	RuntimeExit()
}

func (c *WebDeployService) cmdOSS() {
	cmd := "oss"
	c.AddHelp(cmd, `oss操作`)

	if argsAct != cmd {
		return
	}
	json := JSONService{
		Data: []byte(GetPHPConfigJSON("./data/config/aliyun.php")),
	}
	shellCmd := []string{"ossutil"}
	shellCmd = append(shellCmd, "-e")
	shellCmd = append(shellCmd, strings.Replace(json.GetString("oss", "endpoint"), "-internal", "", -1))
	shellCmd = append(shellCmd, "-i")
	shellCmd = append(shellCmd, json.GetString("access_key"))
	shellCmd = append(shellCmd, "-k")
	shellCmd = append(shellCmd, json.GetString("access_secret"))
	bucket := json.GetString("oss", "bucket")
	bucketPath := "oss://" + bucket

	ossCmd := "ls"
	if len(os.Args) >= 3 {
		ossCmd = os.Args[2]
	}

	if ossCmd == "ls" {
		lsPath := "/"
		if len(os.Args) >= 4 {
			lsPath = os.Args[3]
		}
		lsPath = TrimChar(lsPath, "/")
		if lsPath != "" {
			lsPath += "/"
		}
		shellCmd = append(shellCmd, "ls")
		shellCmd = append(shellCmd, "-sd")
		shellCmd = append(shellCmd, bucketPath+"/"+lsPath)
	} else if ossCmd == "upload" {
		shellCmd = append(shellCmd, "cp")
		shellCmd = append(shellCmd, "-ru")
		shellCmd = append(shellCmd, os.Args[3])
		shellCmd = append(shellCmd, bucketPath+os.Args[4])
	} else if ossCmd == "download" {
		shellCmd = append(shellCmd, "cp")
		shellCmd = append(shellCmd, "-rf")
		shellCmd = append(shellCmd, bucketPath+os.Args[3])
		shellCmd = append(shellCmd, "./")
	} else if ossCmd == "help" {
		println(`
列举文件：
	ls /
上传文件：
	upload ./demo.txt /data/
下载文件:
	download /data/demo.txt
`)
		RuntimeExit()
	} else {
		RuntimeExit()
	}

	content := ExecuteQuiet(shellCmd)
	content = strings.ReplaceAll(content, bucketPath, "")
	println(content)
	RuntimeExit()
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
// GetSLBList 获取负载均衡列表
func (c *AliyunHelper) GetSLBList() []slb.LoadBalancer {
	client, err := slb.NewClientWithAccessKey("cn-shanghai", c.Config.AccessKey, c.Config.AccessSecret)

	request := slb.CreateDescribeLoadBalancersRequest()
	request.Scheme = "https"

	request.PageNumber = requests.NewInteger(100)

	response, err := client.DescribeLoadBalancers(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	return response.LoadBalancers.LoadBalancer
}

// GetECSList 返回ecs列表
func (c *AliyunHelper) GetECSList() []ecs.Instance {
	client, err := ecs.NewClientWithAccessKey("cn-guangzhou", c.Config.AccessKey, c.Config.AccessSecret)

	request := ecs.CreateDescribeInstancesRequest()
	request.Scheme = "https"

	request.PageSize = requests.NewInteger(100)

	response, err := client.DescribeInstances(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	return response.Instances.Instance
}

// RenderEcsList 输出ecs列表
func (c *AliyunHelper) RenderEcsList(list []ecs.Instance) {
	data := [][]string{}
	for _, inst := range list {
		data = append(data, []string{inst.HostName,
			inst.PublicIpAddress.IpAddress[0],
			IntToString(inst.Cpu) + "核/" + IntToString(inst.Memory/1024) + "G",
			IntToString(inst.InternetMaxBandwidthOut) + "M"})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"别名", "IP", "硬件", "宽带"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
