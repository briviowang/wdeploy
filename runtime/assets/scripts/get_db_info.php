<?php
error_reporting(E_ERROR);
$db_config_path = "{{.db_config_path}}";
$content        = [];

if (file_exists($db_config_path)) {
    $content = include($db_config_path);
}
print_r(json_encode($content, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));