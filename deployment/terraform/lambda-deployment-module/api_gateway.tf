data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_api_gateway_rest_api" "dyndns-server" {
  name        = "dyndns-server_api"
  description = "API Gateway for IP PLZ Lambda"
}

resource "aws_api_gateway_resource" "dyndns-server" {
  rest_api_id = aws_api_gateway_rest_api.dyndns-server.id
  parent_id   = aws_api_gateway_rest_api.dyndns-server.root_resource_id
  path_part   = "update"
}

resource "aws_api_gateway_method" "dyndns-server" {
  rest_api_id   = aws_api_gateway_rest_api.dyndns-server.id
  resource_id   = aws_api_gateway_resource.dyndns-server.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id            = aws_api_gateway_rest_api.dyndns-server.id
  stage_name             = aws_api_gateway_stage.dyndns-server_v1.stage_name
  method_path            = "*/*"

  settings {
    logging_level = "OFF"
    throttling_rate_limit  = 1
    throttling_burst_limit = 5
  }
}

resource "aws_api_gateway_integration" "dyndns-server" {
  rest_api_id             = aws_api_gateway_rest_api.dyndns-server.id
  resource_id             = aws_api_gateway_resource.dyndns-server.id
  http_method             = aws_api_gateway_method.dyndns-server.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.dyndns-server.invoke_arn
  timeout_milliseconds    = 5000
}

resource "aws_api_gateway_deployment" "dyndns-server_v1" {
  depends_on = [
    aws_api_gateway_integration.dyndns-server,
    aws_api_gateway_method.dyndns-server
  ]
  rest_api_id = aws_api_gateway_rest_api.dyndns-server.id
}

resource "aws_api_gateway_stage" "dyndns-server_v1" {
  depends_on = [
    aws_api_gateway_integration.dyndns-server,
    aws_api_gateway_method.dyndns-server
  ]
  deployment_id = aws_api_gateway_deployment.dyndns-server_v1.id
  rest_api_id   = aws_api_gateway_rest_api.dyndns-server.id
  stage_name    = "v1"
}

resource "aws_lambda_permission" "api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.dyndns-server.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "arn:aws:execute-api:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.dyndns-server.id}/*/${aws_api_gateway_method.dyndns-server.http_method}${aws_api_gateway_resource.dyndns-server.path}"
}
