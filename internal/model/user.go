/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-03 02:50:18
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 04:20:59
 */
package model

type User struct {
	Account  string `json:"account"`
	Eamil    string `json:"email"`
	PassWord string `json:"passwd"`
}

func (ths *User) Register() (User, error) {
	return *ths, nil
}
