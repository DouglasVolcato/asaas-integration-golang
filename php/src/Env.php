<?php

namespace AsaasIntegration;

class Env
{
    public static function load(string $path): void
    {
        if (!is_file($path)) {
            return;
        }

        $lines = file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
        if ($lines === false) {
            return;
        }

        foreach ($lines as $line) {
            if (str_starts_with(trim($line), '#')) {
                continue;
            }
            [$name, $value] = array_pad(explode('=', $line, 2), 2, '');
            $name = trim($name);
            $value = trim($value, " \t\"'\r\n");
            if ($name === '') {
                continue;
            }
            if (getenv($name) === false) {
                putenv(sprintf('%s=%s', $name, $value));
                $_ENV[$name] = $value;
                $_SERVER[$name] = $value;
            }
        }
    }
}
