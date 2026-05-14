{{ config(
    materialized='table'
) }}

WITH base AS (
    SELECT
    *
    FROM {{ ref('stg_sales_orders') }}
),

customers AS (
    SELECT
    *
    FROM {{ ref('stg_sales_customers') }}
)

SELECT
    base.order_id AS order_id,
    base.customer_id AS customer_id,
    base.amount AS amount,
    base.status AS status,
    customers.full_name AS customer_name,
    customers.email AS customer_email
FROM base
LEFT JOIN customers
  ON base.customer_id = customers.customer_id
