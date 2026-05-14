{{ config(
    materialized='view'
) }}

WITH base AS (
    SELECT
    *
    FROM {{ source('raw_sales', 'orders') }}
)

SELECT
    base.order_id AS order_id,
    base.customer_id AS customer_id,
    base.amount AS amount,
    base.status AS status
FROM base
