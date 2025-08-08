# go-payment-authorizer
worklod for POC purposes. It is simulate a card payment synch

# usecase
    AddPaymentToken

    "step_process":
        {
            "step_process": "AUTHORIZATION-GRPC:STATUS:PENDING",
            "processed_at": "2025-06-02T00:50:40.102903847Z"
        },
        {
            "step_process": "LIMIT:BREACH_LIMIT:CREDIT",
            "processed_at": "2025-06-02T00:50:40.271713828Z"
        },
        {
            "step_process": "ACCOUNT-FROM:OK",
            "processed_at": "2025-06-02T00:50:40.345516442Z"
        },
        {
            "step_process": "LEDGER:WITHDRAW:OK",
            "processed_at": "2025-06-02T00:50:40.616321964Z"
        },
        {
            "step_process": "CARD-ATC:OK",
            "processed_at": "2025-06-02T00:50:40.700784729Z"
        },
        {
            "step_process": "AUTHORIZATION-GRPC:STATUS:OK",
            "processed_at": "2025-06-02T00:50:40.704779203Z"
        }

# table
    table payment
    
    id    |fk_card_id|card_number    |fk_terminal_id|terminal|card_type|card_model|payment_at                   |mcc |status               |currency|amount|request_id                          |transaction_id                      |fraud|created_at                   |updated_at                   |
    ------+----------+---------------+--------------+--------+---------+----------+-----------------------------+----+---------------------+--------+------+------------------------------------+------------------------------------+-----+-----------------------------+-----------------------------+
    192|        35|111.111.111.500|             1|TERM-1  |CREDIT   |CHIP      |2025-04-29 16:11:59.289 -0300|FOOD|AUTHORIZATION-GRPC:OK|BRL     |  22.0|626b7d82-587a-4c9e-b1cf-b9cab2efbe9b|b30d8aea-32be-480e-8c1d-afe01f9dbdf3|     |2025-04-29 16:11:59.289 -0300|2025-04-29 16:12:00.347 -0300|
    194|        35|111.111.111.500|             1|TERM-1  |CREDIT   |CHIP      |2025-04-29 21:14:11.539 -0300|FOOD|AUTHORIZATION:OK     |BRL     |  11.0|c773ed08-b853-4bb7-b4ff-a7d5d23d4c60|f643a74a-6071-4462-8240-1acd6d4afbc5|     |2025-04-29 21:14:11.539 -0300|2025-04-29 21:14:12.079 -0300|
    198|       519|111.111.000.155|             1|TERM-1  |CREDIT   |CHIP      |2025-04-29 21:38:00.409 -0300|FOOD|AUTHORIZATION:OK     |BRL     | 124.0|4466f764-0044-495e-8a2d-6c1301de3a0d|e46d7ea0-4f0c-4dd4-8f55-bc30009a9baf|     |2025-04-29 21:38:00.409 -0300|2025-04-29 21:38:00.920 -0300|
    199|       595|111.111.000.231|             1|TERM-1  |CREDIT   |CHIP      |2025-04-29 21:38:44.476 -0300|FOOD|AUTHORIZATION:OK     |BRL     |  99.0|2958b570-a55c-4aa4-a158-fd6baf46626b|034b8b9c-d35b-4e55-a6b8-5f43a0d183ec|     |2025-04-29 21:38:44.476 -0300|2025-04-29 21:38:44.903 -0300|
    200|       664|111.111.000.300|             1|TERM-1  |CREDIT   |CHIP      |2025-04-29 21:41:00.776 -0300|FOOD|AUTHORIZATION:OK     |BRL     | 141.0|b61d1c2f-a567-47eb-b8c3-0e95a6e7dc78|ff7d9581-fd6a-4554-9e08-2a6db38523cc|     |2025-04-29 21:41:00.776 -0300|2025-04-29 21:41:01.275 -0300|