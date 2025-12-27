<?php

namespace App;

use PDO;
use RuntimeException;

class Database
{
    public static function connect(string $databaseUrl): PDO
    {
        $parts = parse_url($databaseUrl);
        if ($parts === false || !isset($parts['scheme'])) {
            throw new RuntimeException('DATABASE_URL inválida');
        }
        $scheme = $parts['scheme'];
        if ($scheme !== 'postgres' && $scheme !== 'postgresql') {
            throw new RuntimeException('Apenas PostgreSQL é suportado');
        }

        $host = $parts['host'] ?? 'localhost';
        $port = $parts['port'] ?? 5432;
        $user = $parts['user'] ?? '';
        $pass = $parts['pass'] ?? '';
        $dbname = ltrim($parts['path'] ?? '', '/');

        $dsn = "pgsql:host={$host};port={$port};dbname={$dbname}";
        $pdo = new PDO($dsn, $user, $pass, [
            PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        ]);
        return $pdo;
    }
}
