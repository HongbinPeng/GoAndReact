import React from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { updateName } from './store/actions'

export default function User() {
  const user = useSelector(state => state.user)
  const dispatch = useDispatch()
  console.log('User')
  return (
    <div>
      <p>{user.name}</p>
      <p>{user.age}</p>
      <button onClick={() => dispatch(updateName('xxx'))}>更新名字</button>
    </div>
  )
}
