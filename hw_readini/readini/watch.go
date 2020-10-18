package readini

import "fmt"

// Listener 接口
type Listener interface {
	listen(infile string)
}

// ListenFunc 实现接口的函数类型
type ListenFunc func(infile string) (*config, error)

// 实现了接口里的 listen 函数
func (f ListenFunc) listen(infile string) (*config, error) {
	return f(infile)
}

// 监听文件的变化
func isChanged(infile string) (*config, error) {
	// 原始配置文件
	c := new(config)
	c.filePath = infile
	err := c.readini()
	if err != nil {
		return nil, err
	}
	// fmt.Println("hello  " + c.info["server"]["protocol"])
	for {
		// 新的配置文件
		newc := new(config)
		newc.filePath = infile
		err := newc.readini()
		if err != nil {
			return nil, err
		}

		// 判断是否改变
		var change bool = false
		for s1, m1 := range c.info {
			for k1, v1 := range m1 {
				_, ok := newc.info[s1]
				if !ok {
					fmt.Println("File is changed!")
					change = true
				}
				if change == true {
					break
				}

				v2, ok := newc.info[s1][k1]
				if !ok {
					fmt.Println("File is changed!")
					change = true
				}
				if v2 != v1 {
					fmt.Println("File is changed!")
					change = true
				}
				if change == true {
					break
				}
			}
			if change == true {
				break
			}
		}

		c = newc // 每次监听后都需要重新对原始配置文件进行修改
	}
	// return nil, nil
}

// Watch 监听配置文件的变化
func Watch(filename string, listener ListenFunc) (*config, error) {
	listener = isChanged
	return listener.listen(filename)
}
