import { combineReducers } from 'redux'
import { INCREMENT, DECREMENT, UPDATENAME } from './actionTypes'
function counter(state = 10, action) {
  switch (action.type) {
    case INCREMENT:
      return state + action.payload
    case DECREMENT:
      return state - action.payload
    default:
      return state
  }
}

function user(state = { name: 'ifer', age: 18 }, action) {
  switch (action.type) {
    case UPDATENAME:
      return {
        ...state,
        name: action.payload,
      }
    default:
      return state
  }
}

export default combineReducers({
  counter,
  user,
})