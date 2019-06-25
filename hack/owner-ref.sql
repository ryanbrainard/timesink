SELECT
    concat(p.raw -> 'data' -> 'kind', p.raw -> 'data' -> 'metadata' -> 'name'),
    last(concat(c.raw -> 'data' -> 'kind', c.raw -> 'data' -> 'metadata' -> 'name'), c.time),
    c.time
FROM cloud_events AS p
         LEFT JOIN cloud_events AS c ON (
            c.raw -> 'data' -> 'metadata' -> 'ownerReferences' -> 0 -> 'name'       = p.raw -> 'data' -> 'metadata' -> 'name'
        AND c.raw -> 'data' -> 'metadata' -> 'ownerReferences' -> 0 -> 'kind'       = p.raw -> 'data' -> 'kind'
        AND c.raw -> 'data' -> 'metadata' -> 'ownerReferences' -> 0 -> 'apiVersion' = p.raw -> 'data' -> 'apiVersion'
    )
where  c.subject like '%busybox-2%'
GROUP BY 1, 3
ORDER BY 1, 3
