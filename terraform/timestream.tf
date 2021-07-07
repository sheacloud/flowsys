resource "aws_timestreamwrite_database" "flowsys" {
  database_name = "flowsys"
}

resource "aws_timestreamwrite_table" "flows" {
  database_name = aws_timestreamwrite_database.flowsys.database_name
  table_name    = "flows"

  retention_properties {
    memory_store_retention_period_in_hours  = 8
    magnetic_store_retention_period_in_days = 30
  }
}
