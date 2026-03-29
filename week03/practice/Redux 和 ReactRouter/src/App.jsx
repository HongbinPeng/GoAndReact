import React from 'react'
import { increment, decrementAsync } from './store/actions'
import { useSelector, useDispatch } from 'react-redux'
import Test from './Test'
import User from './User'
import { Link, Outlet } from 'react-router'

export default function App() {
  const count = useSelector(state => state.counter)
  const dispatch = useDispatch()
  return (
    <div>
      <p>{count}</p>
      <button onClick={() => dispatch(increment(1))}>+1</button>
      <button onClick={() => dispatch(decrementAsync(2))}>async -2</button>
      <Test/>
      <User/>

      <nav>
        <Link to="/">Home</Link>
        <Link to="/list">List</Link>
        <Link to="/about">About</Link>
      </nav>
      <Outlet/>
    </div>
  )
}
