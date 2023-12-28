SERVER_PATH='{{.ServerPath}}'
cd $SERVER_PATH

# tp3、tp5
for i in ./data/config/db.php ./data/config/database.php; do
    if [[ -f $i ]];then
        php -r "print_r(json_encode(include('$i'), JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));"
        exit
    fi
done

# tp6
for i in ./.env; do
    if [[ -f $i ]];then
        php -r '
$env_file_path = "'$i'";
$content       = [];
if (file_exists($env_file_path)) {
    $res     = parse_ini_file($env_file_path, true);
    $content = parse_array($res)["database"];
}
print_r(json_encode($content, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));

function parse_array($data)
{
    if (!is_array($data)) {
        return $data;
    }
    $result = [];
    foreach ($data as $key => $val) {
        $result[strtolower($key)] = parse_array($val);
    }
    return $result;
}
'
        exit
    fi
done

echo "{}"
