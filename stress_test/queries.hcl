query "get_series" {
    sql = "SELECT md5(random()::text) AS hash, * FROM generate_series(1, $1) ORDER BY RANDOM()"
    timeout = 10
}