unit ApiClient;

<?=$config['copyright']?>

interface
uses
  System.Classes, 
  System.Generics.Collections,
  System.JSON,
  System.Net.HttpClient, 
  System.Net.HttpClientComponent, 
  System.Net.UrlClient,
  System.SysUtils;

type
  TApiBase = class(TObject)
  public
    function JString(data: TJSONObject; key: string): string;
    function JObject(data: string): TJSONObject;
  end;

<?foreach($entity_list as $info){?>
  T<?=$info['name']?>=class;
<?}?>

<?foreach($entity_list as $info){?>
  T<?=$info['name']?> = class(TApiBase)
  public
<?foreach($info['fields'] as $key=>$val){?>
<?if(strpos($key,'?')>0){
    $hasList=trim($key,'?');
    ?>
<?if($val=='String'){?>
    <?=getField($hasList)?>:TStringList;
<?}else{?>
    <?=getField($hasList)?>:TList;
<?}?>
<?}elseif($val=='String'){?>
<?if(isset($info['comments'][$key])){?>
    { <?=$info['comments'][$key]?> }
<?}?>
    <?=getField($key);?>:String;
<?}else{?>
    <?=getField($key);?>:T<?=$val;?>;
<?}?>
<?}?>

<?if($info['type']!='request'){?>
    constructor Create(AData:string);
<?}?>
    function ToJson():TJsonObject;
    function ToString():string;override;
  end;

<?}?>

  TApiRequestSuccess =reference to procedure(AData: string);
  TApiRequestError =reference to procedure(AData: string);
  TApiRequestCallback=reference to procedure(AData: TJSONObject;AUrl:string);

  TApiClient = class(TObject)
  private
    client: TNetHttpClient;
    requst: TNetHTTPRequest;
    baseUrl: string;
    requestExtraParams:TDictionary<string,string>;
    successCallback: TApiRequestSuccess;
    errorCallback: TApiRequestError;
    beforeRequestCallback:TApiRequestCallback;
    afterRequestCallback:TApiRequestCallback;
    procedure post(AUrl: string; AParams: TJsonObject);
    procedure defaultProcedure(AData:string);
    procedure defaultRequestCallback(AData: TJSONObject;AUrl:string);
    procedure apiResponse(const Sender: TObject; const AResponse: IHTTPResponse);
    procedure apiError(const Sender: TObject; const AError: string);
  public
    constructor Create();
    procedure SetProxy(AHost:string;APort:Integer);
    procedure SetBaseUrl(AUrl:String);
    procedure UpdateRequestParams(AParams:TDictionary<string,string>);
    procedure SetBeforeRequestCallback(ACallback:TApiRequestCallback);
    procedure SetAfterRequestCallback(ACallback:TApiRequestCallback);
<?foreach($parseResult['api'] as $key=>$val){?>
    procedure do<?=$key?>(Request: T<?=$key?>Request; success:TApiRequestSuccess; error: TApiRequestError);overload;
    procedure do<?=$key?>(Request: T<?=$key?>Request; success:TApiRequestSuccess);overload;
<?}?>

  end;

implementation

<?foreach($entity_list as $info){?>
<?if($info['type']!='request'){?>
constructor T<?=$info['name']?>.Create(AData:string);
var 
  jsonObject:TJsonObject;
<?$hasList=false;
foreach($info['fields'] as $key=>$val){
    if(strpos($key,'?')>0){
      $hasList=true; 
      $key=trim($key,'?');?>
  <?=$key?>Array:TJSONArray;
<?}}?>
<?if($hasList){?>
  I: Integer;
<?}?>
begin
  jsonObject:=JObject(AData);
<?foreach($info['fields'] as $key=>$val){
    if(strpos($key,'?')>0){
        $key=trim($key,'?');
    ?>

<?if($val=='String'){?>
  <?=getField($key)?>:=TStringList.Create;
<?}else{?>
  <?=getField($key)?>:=TList.Create;
<?}?>
  <?=$key?>Array:=jsonObject.GetValue('<?=$key?>') as TJSONArray;
  if <?=$key?>Array <> nil then
  begin
    for I := 0 to <?=$key?>Array.count - 1 do
    begin
<?if($val=='String'){?>
      <?=getField($key)?>.Add(<?=$key?>Array.Items[i].ToString);
<?}else{?>
      <?=getField($key)?>.Add(T<?=$val?>.Create(<?=$key?>Array.Items[i].ToString));
<?}?>
    end;
  end;

<?}else if($val=="String"){?>
  <?=getField($key);?>:=JString(jsonObject,'<?=$key;?>');
<?}else{?>
  <?=getField($key);?>:=T<?=$val?>.Create(JString(jsonObject,'<?=$key;?>'));
<?}?>
<?}?>
  jsonObject.Free
end;
<?}?>

function T<?=$info['name']?>.ToJson():TJsonObject;
<?if($hasList){?>
var
<?foreach($info['fields'] as $key=>$val){
    if(strpos($key,'?')>0){ 
      $key=trim($key,'?');?>
  <?=$key?>Array:TJSONArray;
<?}}?>
  I: Integer;
<?}?>

begin
  Result:=TJsonObject.Create();

<?foreach($info['fields'] as $key=>$val){
    if(strpos($key,'?')>0){
        $key=trim($key,'?');?>
  <?=$key?>Array := TJSONArray.Create();
  for I := 0 to <?=$key?>Array.count - 1 do
  begin
<?if(in_array($val,['String','Integer'])){?>
    <?=$key?>Array.Add(<?=getField($key)?>[i]);
<?}else{?>
    <?=$key?>Array.Add(T<?=$val?>(<?=getField($key)?>.Items[i]).ToJson());
<?}?>
  end;
  Result.AddPair(TJSONPair.Create('<?=$key;?>',<?=$key?>Array));     

<?}elseif($val=="String"){?>
  if <?=getField($key);?> <> '' then
    Result.AddPair(TJSONPair.Create('<?=$key;?>',<?=getField($key);?>));
<?}else{?>
  if <?=getField($key);?> <> nil then
    Result.AddPair(TJSONPair.Create('<?=$key;?>',<?=getField($key);?>.ToJson()));
<?}?>
<?}?>
end;

function T<?=$info['name']?>.ToString():string;
begin
  Result:=ToJson().ToString();
end;
  
<?}?>

function TApiBase.JObject(data: string): TJSONObject;
begin
  if ( data = '' ) or ( data = '[]' ) then
  begin
    data:='{}'
  end;
  Result := TJSONObject.ParseJSONValue(data) as TJSONObject;
end;

function TApiBase.JString(data: TJSONObject; key: string): string;
begin
  if data = nil then
  begin
    Result := '';
    Exit
  end;

  if data.GetValue(key) <> nil then
  begin
    Result := data.GetValue(key).ToString;
  end
  else
  begin
    Result := '';
  end;
end;

constructor TApiClient.Create();
begin
  requestExtraParams:=TDictionary<string,string>.Create;
  beforeRequestCallback:=defaultRequestCallback;
  afterRequestCallback:=defaultRequestCallback;

  client := TNetHTTPClient.Create(nil);
  client.UserAgent:='delphi_sdk';

  requst := TNetHTTPRequest.Create(nil);
  requst.Asynchronous := true;
  requst.Client := client;
  requst.OnRequestCompleted := apiResponse;
  requst.OnRequestError := apiError;
end;

procedure TApiClient.SetProxy(AHost:string;APort:Integer);
var 
  proxy:TProxySettings;
begin
  proxy:=TProxySettings.Create(AHost,APort);
  client.ProxySettings :=proxy;
end;

procedure TApiClient.SetBaseUrl(AUrl:String);
begin
  baseUrl := AUrl;
end;

procedure TApiClient.UpdateRequestParams(AParams:TDictionary<string,string>);
begin
  requestExtraParams:=AParams;
end;

procedure TApiClient.SetBeforeRequestCallback(ACallback:TApiRequestCallback);
begin
  beforeRequestCallback:=ACallback;
end;

procedure TApiClient.SetAfterRequestCallback(ACallback:TApiRequestCallback);
begin
  afterRequestCallback:=ACallback;
end;

procedure TApiClient.post(AUrl: string; AParams: TJsonObject);
var
  requestParams: TStringList;
  k:String;
begin
  beforeRequestCallback(AParams,AUrl);

  requestParams := TStringList.Create;
  for k in requestExtraParams.Keys.ToArray do
  begin
    requestParams.AddPair(k,requestExtraParams.Items[k]);
  end;
  
  requestParams.AddPair('platform', 'windows');
  requestParams.AddPair('data',Aparams.ToString);
  requst.Post(AUrl, requestParams);
end;

procedure TApiClient.apiResponse(const Sender: TObject; const AResponse:
  IHTTPResponse);
var
  content: string;
  jsonObject: TJsonObject;
  status: string;
begin
  content := AResponse.ContentAsString;
  afterRequestCallback(jsonObject,requst.URL);

  jsonObject := TJSONObject.ParseJSONValue(content) as TJSONObject;

  status := jsonObject.GetValue('status').ToString;
  if status = '1' then
  begin
    successCallback(content);
  end
  else
  begin
    errorCallback(content);
  end;
end;

procedure TApiClient.apiError(const Sender: TObject; const AError: string);
var
  content: string;
begin
  content := AError;
  afterRequestCallback(TJSONObject.Create,requst.URL);
end;

procedure TApiClient.defaultProcedure(AData:string);
begin
  //  
end;

procedure TApiClient.defaultRequestCallback(AData: TJSONObject;AUrl:string);
begin
  //  
end;

<?foreach($parseResult['api'] as $key=>$val){?>
<?if($val['comment']){?>
{
<?=$val['comment']?> 
}
<?}?>
procedure TApiClient.do<?=$key?>(Request: T<?=$key?>Request; success:TApiRequestSuccess);
begin
  do<?=$key?>(Request,success,defaultProcedure);
end;

procedure TApiClient.do<?=$key?>(Request: T<?=$key?>Request;success: TApiRequestSuccess; error: TApiRequestError);
begin
  successCallback := success;
  errorCallback := error;
  post(baseUrl + '/<?=$val['url']?>', Request.toJSON);
end;
<?}?>

end.