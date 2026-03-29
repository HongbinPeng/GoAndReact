import React from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { increment } from './store/actions'

export default React.memo(function Test() {
  const count = useSelector(state => state.counter)
  const dispatch = useDispatch()
  console.log('Test')
  return (
    <div>
      <p>{count}</p>
      <button onClick={() => dispatch(increment(3))}>+1</button>
    </div>
  )
})
