// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a written permission.

package commitmsg

type Strategy func(message string) Category

//var DefaultStrategies = []Strategy{
//	GetCategoryByKeywords,
//}

type Classification struct {
	KeywordClassification Category
}

func Classify(message string) Classification {
	k := GetCategoryByKeywords(message)

	return Classification{
		KeywordClassification: k,
	}
}

//func GetCategory(message string) Category {
//	for _, strategy := range DefaultStrategies {
//		category := strategy(message)
//		if category != Others {
//			return category
//		}
//	}
//	return Others
//}
