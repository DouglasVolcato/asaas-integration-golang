<?php

namespace App;

class Env
{
    public static function load(?string $path = null): void
    {
        $candidates = [];
        if ($path !== null) {
            $candidates[] = $path;
        }
        $candidates[] = dirname(__DIR__) . '/.env';
        $candidates[] = dirname(__DIR__, 2) . '/.env';

        foreach ($candidates as $file) {
            if ($file && is_readable($file)) {
                self::parse($file);
                break;
            }
        }
    }

    private static function parse(string $file): void
    {
        $lines = file($file, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
        if ($lines === false) {
            return;
        }
        foreach ($lines as $line) {
            if (str_starts_with(trim($line), '#')) {
                continue;
            }
            if (!str_contains($line, '=')) {
                continue;
            }
            [$key, $value] = array_map('trim', explode('=', $line, 2));
            if ($key !== '') {
                putenv("{$key}={$value}");
                $_ENV[$key] = $value;
                $_SERVER[$key] = $value;
            }
        }
    }
}
