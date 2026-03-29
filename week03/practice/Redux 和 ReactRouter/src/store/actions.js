import { INCREMENT, DECREMENT, UPDATENAME } from './actionTypes'

export const increment = (payload) => ({
  type: INCREMENT,
  payload,
})

export const decrement = (payload) => ({
  type: DECREMENT,
  payload,
})

export const updateName = (payload) => ({
  type: UPDATENAME,
  payload,
})

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms))

export const decrementAsync = (payload) => {
  return async (dispatch) => {
    await sleep(2000)
    dispatch(decrement(payload))
  }
}