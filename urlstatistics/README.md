# README

## output
text:
```txt
+---------------------------------------------------+------------+------------------+------------------+------------------+------------------+------------------+------------------+
| Request URL                                       | Method     | Times            | Status Times     | Total Used(s)    | Max Used(μs)     | Min Used(μs)     | Avg Used(μs)     |
+---------------------------------------------------+------------+------------------+------------------+------------------+------------------+------------------+------------------+
| /docs/openapi-ui                                  | GET        |                1 | 200:1            |         0.000000 |         0.228000 |         0.228000 |         0.228000 |
| /a                                                | GET        |                2 | 200:1, 201:1     |         0.000003 |         1.908000 |         1.019000 |         1.463000 |
| /docs/openapi                                     | POST       |                1 | 200:1            |         0.000000 |         0.240000 |         0.240000 |         0.240000 |
+---------------------------------------------------+------------+------------------+------------------+------------------+------------------+------------------+------------------+
```

json:
```json
[
    {
        "url":"/docs/openapi-ui",
        "method":"GET",
        "times":1,
        "total_used":0.000000228,
        "max_used":0.228,
        "min_used":0.228,
        "avg_used":0.228,
        "codes":[
            {
                "code":200,
                "count":1
            }
        ]
    },
    {
        "url":"/a",
        "method":"GET",
        "times":2,
        "total_used":0.000002927,
        "max_used":1.908,
        "min_used":1.019,
        "avg_used":1.463,
        "codes":[
            {
                "code":200,
                "count":1
            },
            {
                "code":201,
                "count":1
            }
        ]
    },
    {
        "url":"/docs/openapi",
        "method":"POST",
        "times":1,
        "total_used":0.00000024,
        "max_used":0.24,
        "min_used":0.24,
        "avg_used":0.24,
        "codes":[
            {
                "code":200,
                "count":1
            }
        ]
    }
]
```