package sql

import (
	"main/internal/global"
	"main/internal/utils"
)

// 获取指定题目的测试样例
func SelectTestCasesByPid(pid int) []global.TestCaseRequest {
	var testCases []global.TestCaseRequest
	utils.ConnectSql().Table("test_cases").Where("pid = ?", pid).Select("input_data,output_data").Find(&testCases)
	return testCases
}

// 修改提交记录状态
func ModifyJudgeStatus(Uid, Pid int, Result string) bool {
	// 将result为Running...的记录修改为返回状态
	err := utils.ConnectSql().Table("submit_records").Where("uid = ? AND pid = ? AND result = ?", Uid, Pid, "Running...").Update("result", Result)
	return err == nil
}

// 获取指定题目信息
func SelectProblemByPid(pid int) global.Problem {
	var problem global.Problem
	utils.ConnectSql().Table("problems").Where("pid = ?", pid).Select("pid, difficulty, title, content, timelimit, memorylimit, input, output, contest_id, is_visible").Find(&problem)
	return problem
}
