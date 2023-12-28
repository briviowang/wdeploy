package {{.package_prefix}};
{{.copyright}}
import android.annotation.TargetApi;
import android.app.Activity;
import android.content.pm.PackageInfo;
import android.os.AsyncTask;
import android.os.Build;
import android.util.Log;

import {{.package_prefix}}.request.*;

import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.Map;

public class ApiClient {
    protected boolean showProgress = true;
    protected OnApiClientListener apiClientListener;
    private ArrayList<RequestTask> requestTasks;
    private Activity context;

    public ApiClient(Activity context, OnApiClientListener apiClientListener) {
        this.context = context;
        this.apiClientListener = apiClientListener;
    }

    public ApiClient hideProgress() {
        showProgress = false;
        return this;
    }

    public interface OnApiClientListener {
        void beforeAjaxPost(JSONObject requestData, String url, boolean showProgress);

        boolean afterAjaxPost(JSONObject responseData, String url, boolean showProgress);

        String getToken();

        String getApiUrl();
    }

    public interface OnSuccessListener {
        void callback(JSONObject object);
    }

    public interface OnFailListener {
        void callback(JSONObject object);
    }

    public class RequestCallback {
        public String url;
        public String token;
        public String userAgent;
        public JSONObject requestData;
        public JSONObject responseData;
        public OnSuccessListener onSuccessListener;
        public OnFailListener onFailListener;
        public boolean showProgress;
    }

    protected Map<String, String> extraRequestParams;

    /**
     * 增加额外POST参数，其中key不能是token、bundle、platform、data，会被忽略
     */
    public void updateRequestParams(Map<String, String> params) {
        this.extraRequestParams = params;
    }

    @TargetApi(Build.VERSION_CODES.CUPCAKE)
    class RequestTask extends AsyncTask<Void, Integer, String> {
        private RequestCallback callback;

        public RequestTask setCallback(RequestCallback callback) {
            this.callback = callback;
            return this;
        }

        @Override
        protected void onPreExecute() {
            apiClientListener.beforeAjaxPost(callback.requestData, callback.url, callback.showProgress);
            super.onPreExecute();
        }

        @Override
        protected void onPostExecute(String s) {
            if (apiClientListener.afterAjaxPost(callback.responseData, callback.url, callback.showProgress)) {
                boolean returnSuccess;
                try {
                    returnSuccess = callback.responseData.getInt("status") == 1;
                } catch (Exception e) {
                    returnSuccess = false;
                }
                try {
                    if (returnSuccess) {
                        callback.onSuccessListener.callback(callback.responseData);
                    } else {
                        callback.onFailListener.callback(callback.responseData);
                    }
                } catch (Exception e) {
                    e.printStackTrace();
                }
                super.onPostExecute(s);
            }
        }

        @Override
        protected String doInBackground(Void... params) {
            String result = "";
            try {
                Log.d("api_sdk_curl", String.format("curl %s -d 'platform=android&token=%s&data=%s'",
                        callback.url,
                        callback.token == null ? "" : callback.token,
                        callback.requestData.toString()
                ));

                Map<String, Object> postParams = new LinkedHashMap<>();
                postParams.put("token", callback.token);
                
                if (extraRequestParams != null) {
                    for (Map.Entry<String, String> entry : extraRequestParams.entrySet()) {
                        String key = entry.getKey();
                        postParams.put(key, entry.getValue());
                    }
                }
                //设置通用参数
                postParams.put("platform", "android");
                postParams.put("bundle", context.getPackageName());
                PackageInfo packageInfo = context.getPackageManager().getPackageInfo(context.getPackageName(), 0);
                postParams.put("version", packageInfo.versionName + "." + packageInfo.versionCode);
                postParams.put("data", callback.requestData.toString());

                //data编码
                StringBuilder postData = new StringBuilder();
                for (Map.Entry<String, Object> param : postParams.entrySet()) {
                    if (postData.length() != 0) postData.append('&');
                    postData.append(URLEncoder.encode(param.getKey(), "UTF-8"));
                    postData.append('=');
                    postData.append(URLEncoder.encode(String.valueOf(param.getValue()), "UTF-8"));
                }
                byte[] postDataBytes = postData.toString().getBytes("UTF-8");

                //执行post请求
                URL url = new URL(callback.url);
                HttpURLConnection conn = (HttpURLConnection) url.openConnection();
                //设置user agent
                if (callback.userAgent != null) {
                    conn.setRequestProperty("User-Agent", callback.userAgent);
                }

                conn.setRequestMethod("POST");
                conn.setRequestProperty("Content-Type", "application/x-www-form-urlencoded");
                conn.setRequestProperty("Content-Length", String.valueOf(postDataBytes.length));
                conn.setDoOutput(true);
                conn.getOutputStream().write(postDataBytes);

                BufferedReader in = new BufferedReader(new InputStreamReader(conn.getInputStream(), "UTF-8"));

                StringBuilder sb = new StringBuilder();
                for (String line; (line = in.readLine()) != null; ) {
                    sb.append(line).append('\n');
                }
                result = sb.toString();
                callback.responseData = toJSONObject(result);
            } catch (Exception e) {
                e.printStackTrace();
            }
            return result;
        }
    }

    private void httpPost(String url, String data, OnSuccessListener successListener, OnFailListener failListener) {
        RequestCallback callback = new RequestCallback();
        callback.url = apiClientListener.getApiUrl() + url;
        callback.token = apiClientListener.getToken();
        callback.requestData = toJSONObject(data);
        if (successListener == null) {
            callback.onSuccessListener = new OnSuccessListener() {
                @Override
                public void callback(JSONObject object) {

                }
            };
        } else {
            callback.onSuccessListener = successListener;
        }
        if (failListener == null) {
            callback.onFailListener = new OnFailListener() {
                @Override
                public void callback(JSONObject object) {

                }
            };
        } else {
            callback.onFailListener = failListener;
        }

        callback.showProgress = showProgress;

        RequestTask task = new RequestTask();
        if (requestTasks == null) {
            requestTasks = new ArrayList<>();
        }
        requestTasks.add(task);

        task.setCallback(callback).execute();
    }

    public void cancelRequests() {
        try {
            for (RequestTask task : requestTasks) {
                task.cancel(true);
            }
        } catch (Exception e) {
        } finally {
            requestTasks = new ArrayList<>();
        }
    }

    private JSONObject toJSONObject(String data) {
        JSONObject result;
        try {
            result = new JSONObject(data);
        } catch (Exception e) {
            result = new JSONObject();
        }
        return result;
    }
{{range $key,$val:=.apiList}}
{{if $val.comment }}
{{$val.comment}}
{{end}}
    public static final String {{$key}} ="/{{$val.url}}";
    public void do{{$key}}({{$.prefix}}{{$key}}Request request,OnSuccessListener successListener){
        do{{$key}}(request,successListener,null);
    }
    public void do{{$key}}({{$.prefix}}{{$key}}Request request,OnSuccessListener successListener,OnFailListener failListener) {
        httpPost({{$key}}, request.toString(),successListener,failListener);
    }
{{end}}
}
