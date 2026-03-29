// store: 整个数据的仓库，负责关联 reducer 和 action，通过 store 对象可以给 reducer 分配 action
import { legacy_createStore, applyMiddleware } from 'redux'
import { composeWithDevTools } from '@redux-devtools/extension'
import reducer from './reducers'
import { thunk } from 'redux-thunk'
const store = legacy_createStore(reducer, composeWithDevTools(applyMiddleware(thunk)))
export default store