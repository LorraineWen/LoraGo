package lora_router

/*
*@Author: LorraineWen
*@Date: 2025/2/23 14:23:49
*该文件主要实现前缀树，支持动态路由和通配符
*还是存在缺陷，不能同时支持/getname/:id和/getname/*
 */
import "strings"

type trieNode struct {
	name       string      //节点名称(user,order,get)
	children   []*trieNode //前缀树的子节点
	routerName string
	isEnd      bool //是否遍历到根节点，避免一种情况，如果注册了/user/hello/amie，那么访问/user/hello同样有效(返回405，而不是404)，但是我们并没有注册/user/hello
}

// 负责放入路径
func (t *trieNode) put(name string) {
	root := t
	strs := strings.Split(name, "/")
	for index, path := range strs {
		if index == 0 { //分隔出来的第一个是空格
			continue
		}
		children := t.children
		isMatch := false
		for _, child := range children {
			if child.name == path {
				isMatch = true
				t = child
				break
			}
		}
		//放入的节点一定都是尾节点
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			child := &trieNode{name: path, children: make([]*trieNode, 0), isEnd: isEnd}
			children = append(children, child)
			t.children = children
			t = child
		}
	}
	t = root
}

// 负责拿出路径，返回group后面的完整路径，是/getname/:id，而不是/getname/1或者/getname/2
func (t *trieNode) get(name string) *trieNode {
	strs := strings.Split(name, "/")
	routerName := ""
	for index, path := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, child := range children {
			if child.name == path || child.name == "*" || strings.Contains(child.name, ":") { //如果同时注册:id和*类型的路由，那么就会出问题
				isMatch = true
				routerName += "/" + child.name
				child.routerName = routerName
				t = child
				if index == len(strs)-1 {
					return child
				}
				break
			}
		}
		if !isMatch {
			for _, child := range children {
				if child.name == "**" {
					routerName += "/" + child.name
					child.routerName = routerName
					return child
				}
			}
		}
	}
	return nil
}
