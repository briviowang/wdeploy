<?php
file_put_contents("{{.json_path}}", json_encode(include("{{.api_config_path}}"), JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT));