<?php
error_reporting(E_ERROR);
$env_file_path = "{{.env_file}}";
$content       = [];

if (file_exists($env_file_path)) {
    $res     = parse_ini_file($env_file_path, true);
    $content = parse_array($res)['database'];
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
