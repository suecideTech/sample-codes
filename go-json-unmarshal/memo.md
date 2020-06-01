# JSONのゆるふわUnmershal

- 参考
  - https://www.kaoriya.net/blog/2016/06/25/
- 空のInterface型にUnmershal結果を取り出し、型アサーションでKeyとValueを推測して扱う
  ```
  json.Unmarshal([]byte(blob), &jsonData)
  jsonKeyValue, ok := jsonData.(map[string]interface{})
  if ok == true {
    // 型アサーションに成功したため"key": "value"の構造
  }
  value := jsonKeyValue["array"]
  jsonArrayValue, ok := value .([]interface{})
  if ok == true {
    // 型アサーションに成功したため"value": [...]の構造
  }
  ``