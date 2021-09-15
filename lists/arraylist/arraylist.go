// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package arraylist implements the array list.
//
// Structure is not thread safe. 非并发安全
//
// Reference: https://en.wikipedia.org/wiki/List_%28abstract_data_type%29
package arraylist

import (
	"fmt"
	"strings"

	"github.com/emirpasic/gods/lists"
	"github.com/emirpasic/gods/utils"
)

// List holds the elements in a slice
// 单向链表，不带头结点
type List struct {
	elements []interface{}
	size     int
}

const (
	growthFactor = float32(2.0)  // growth by 100%
	shrinkFactor = float32(0.25) // shrink when size is 25% of capacity (0 means never shrink)
)

// 用于断言
func assertListImplementation() {
	var _ lists.List = (*List)(nil)
}

// New instantiates a new list and adds the passed values, if any, to the list
// 实例化List，同时可以增加多个values
func New(values ...interface{}) *List {
	list := &List{}
	if len(values) > 0 {
		list.Add(values...)
	}
	return list
}

// Add appends a value at the end of the list
func (list *List) Add(values ...interface{}) {
	list.growBy(len(values))
	for _, value := range values {
		// list.size表示当前个数，又因为索引从0开始
		// 因此新增的一个刚好就是list.size所在索引
		list.elements[list.size] = value
		list.size++
	}
}

// Get returns the element at index.
// Second return parameter is true if index is within bounds of the array and array is not empty, otherwise false.
func (list *List) Get(index int) (interface{}, bool) {
	// 根据索引获取节点数据
	// 先判断索引是否在合法范围
	if !list.withinRange(index) {
		return nil, false
	}
	// 直接根据索引返回数值
	return list.elements[index], true
}

// Remove removes the element at the given index from the list.
func (list *List) Remove(index int) {
	// 先判断索引是否在合法范围
	if !list.withinRange(index) {
		return
	}
	// 将索引位置值置为nil
	list.elements[index] = nil                                    // cleanup reference
	// 将索引之后的值向前拷贝
	copy(list.elements[index:], list.elements[index+1:list.size]) // shift to the left by one (slow operation, need ways to optimize this)
	list.size--
	// 压缩的根本也是申请小数组，然后拷贝到数组中
	list.shrink()
}

// Contains checks if elements (one or more) are present in the set.
// All elements have to be present in the set for the method to return true.
// Performance time complexity of n^2.
// Returns true if no arguments are passed at all, i.e. set is always super-set of empty set.
func (list *List) Contains(values ...interface{}) bool {
    // 判断是否包含，多个中任意一个是否包含，其实就是要遍历
	for _, searchValue := range values {
		found := false
		for _, element := range list.elements {
			if element == searchValue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	// 如果是空参数，那么也返回true
	return true
}

// Values returns all elements in the list.
// 返回值数组，其实就是新建一个切片，拷贝进去
func (list *List) Values() []interface{} {
	newElements := make([]interface{}, list.size, list.size)
	copy(newElements, list.elements[:list.size])
	return newElements
}

//IndexOf returns index of provided element
// 获取对应值所在索引，如果不存在则返回-1
func (list *List) IndexOf(value interface{}) int {
	if list.size == 0 {
		return -1
	}
	for index, element := range list.elements {
		if element == value {
			return index
		}
	}
	return -1
}

// Empty returns true if list does not contain any elements.
// 判断数组列表是否为空
func (list *List) Empty() bool {
	return list.size == 0
}

// Size returns number of elements within the list.
// 获取数组列表
func (list *List) Size() int {
	return list.size
}

// Clear removes all elements from the list.
// 清空数组列表
func (list *List) Clear() {
	list.size = 0
	list.elements = []interface{}{}
}

// Sort sorts values (in-place) using.
// 数组列表排序
func (list *List) Sort(comparator utils.Comparator) {
	// 如果只有一个元素，毫无疑问，已经排序
	if len(list.elements) < 2 {
		return
	}
	utils.Sort(list.elements[:list.size], comparator)
}

// Swap swaps the two values at the specified positions.
// 将两个索引对应数值对换
func (list *List) Swap(i, j int) {
	if list.withinRange(i) && list.withinRange(j) {
		list.elements[i], list.elements[j] = list.elements[j], list.elements[i]
	}
}

// Insert inserts values at specified index position shifting the value at that position (if any) and any subsequent elements to the right.
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List) Insert(index int, values ...interface{}) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(values...)
		}
		return
	}

	l := len(values)
	list.growBy(l)
	list.size += l
	// 留出l个位置
	copy(list.elements[index+l:], list.elements[index:list.size-l])
	// 填充多个values
	copy(list.elements[index:], values)
}

// Set the value at specified index
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
// 指定索引的值修改为value
func (list *List) Set(index int, value interface{}) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(value)
		}
		return
	}

	list.elements[index] = value
}

// String returns a string representation of container
// 所有值转为字符串，逗号分隔
func (list *List) String() string {
	str := "ArrayList\n"
	values := []string{}
	for _, value := range list.elements[:list.size] {
		values = append(values, fmt.Sprintf("%v", value))
	}
	str += strings.Join(values, ", ")
	return str
}

// Check that the index is within bounds of the list
func (list *List) withinRange(index int) bool {
	// 索引从0开始
	return index >= 0 && index < list.size
}

func (list *List) resize(cap int) {
	// 切片扩容，本质就是创建一个新的切片，再拷贝进去
	// 底层仍然是数组
	newElements := make([]interface{}, cap, cap)
	// 内置copy函数，dest = src
	copy(newElements, list.elements)
	list.elements = newElements
}

// Expand the array if necessary, i.e. capacity will be reached if we add n elements
func (list *List) growBy(n int) {
	// When capacity is reached, grow by a factor of growthFactor and add number of elements
	currentCapacity := cap(list.elements)
	if list.size+n >= currentCapacity {
		newCapacity := int(growthFactor * float32(currentCapacity+n))
		list.resize(newCapacity)
	}
}

// Shrink the array if necessary, i.e. when size is shrinkFactor percent of current capacity
// 缩小数组列表
func (list *List) shrink() {
	if shrinkFactor == 0.0 {
		return
	}
	// Shrink when size is at shrinkFactor * capacity
	currentCapacity := cap(list.elements)
	if list.size <= int(float32(currentCapacity)*shrinkFactor) {
		list.resize(list.size)
	}
}
