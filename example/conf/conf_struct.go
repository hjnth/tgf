//Auto generated by tgf util
//created at 2023-04-11 00:01:09.146478 +0800 CST m=+0.010015902

package conf

type ChapterConf struct {

	//唯一id
	Id uint32

	//章节名称
	Name string

	//限制等级
	LimitLevel uint32

	//下一关id
	NextId uint32
}

type HeroConf struct {

	//唯一id
	Id string

	//名字
	Name string

	//攻击
	Attack int32

	//防御
	Defend int32

	//奖励
	Reward []int32
}

type EquiConf struct {

	//唯一id
	Id uint32

	//武器名字
	Name string

	//攻击加成
	Attack int32

	//指定装备英雄id
	HeroId string
}
