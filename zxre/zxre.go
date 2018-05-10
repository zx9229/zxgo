package zxre

/*
查看 golang 的 regexp 的帮助文档的方式:
1. 进入 golang 官网 (https://golang.org/)
2. 点击 Documents
3. 点击 Packages
4. 点击 regexp
5. 查看 Index 和 Examples
*/
import "regexp"

func CalcAllGroupDict(pattern string, content string) []map[string]string {
	all_group_dict_ := make([]map[string]string, 0)

	patternObj := regexp.MustCompile(pattern)
	for _, submatches := range patternObj.FindAllStringSubmatchIndex(content, -1) {
		cur_group_dict_ := map[string]string{}
		for _, key := range patternObj.SubexpNames() {
			if len(key) == 0 {
				continue
			}
			result := []byte{}
			result = patternObj.ExpandString(result, "${"+key+"}", content, submatches)
			cur_group_dict_[key] = string(result)
		}
		all_group_dict_ = append(all_group_dict_, cur_group_dict_)
	}

	if len(all_group_dict_) == 0 {
		all_group_dict_ = nil
	}

	return all_group_dict_
}
