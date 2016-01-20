#blocks
#curl -X POST 'localhost:7071/pattern' -d '[{"alias":"foo","type":"block","spec":"+"},{"type":"block","spec":"+","position":{"x":22,"y":100}}]'

#groups + children
#curl -X POST 'localhost:7071/pattern' -d '[{"type":"block","spec":"concat","id":"3"},{"type":"block","spec":"sink","id":"4"},{"type":"group","children":[{"id":"3"},{"id":"4"}]}]'

curl -X POST 'localhost:7071/pattern' -d '[{"type":"source","spec":"value"}]'
