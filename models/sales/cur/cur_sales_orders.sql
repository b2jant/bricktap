{{ config(
    materialized='incremental',
    unique_key='order_id',
    incremental_strategy='merge'
) }}

WITH base AS (
    SELECT
    *
    FROM {{ ref('int_sales_orders_joined') }}
)

SELECT
    base.order_id AS order_id,
    base.amount AS amount,
    base.status AS status,
    base.customer_name AS customer_name,
    base.is_high_value AS is_high_value
FROM base
