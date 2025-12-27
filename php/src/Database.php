<?php

namespace AsaasIntegration;

use PDO;

class Database
{
    public static function connect(string $databaseUrl): PDO
    {
        $parts = parse_url($databaseUrl);
        if ($parts === false || !isset($parts['scheme'])) {
            throw new \RuntimeException('DATABASE_URL inválida');
        }

        $scheme = $parts['scheme'];
        if ($scheme !== 'postgres' && $scheme !== 'postgresql') {
            throw new \RuntimeException('Apenas PostgreSQL é suportado');
        }

        $host = $parts['host'] ?? 'localhost';
        $port = $parts['port'] ?? 5432;
        $user = $parts['user'] ?? '';
        $pass = $parts['pass'] ?? '';
        $dbname = ltrim($parts['path'] ?? '', '/');

        $dsn = sprintf('pgsql:host=%s;port=%s;dbname=%s', $host, $port, $dbname);
        $pdo = new PDO($dsn, $user, $pass, [
            PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
            PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        ]);

        return $pdo;
    }
}
