import rulesReducer, {
  defaultNotificationRulesState,
} from 'src/alerting/reducers/notifications/rules'
import {
  setAllNotificationRules,
  setRule,
  setCurrentRule,
  removeRule,
} from 'src/alerting/actions/notifications/rules'
import {RemoteDataState} from 'src/types'
import {NEW_RULE_DRAFT} from 'src/alerting/constants'

describe('rulesReducer', () => {
  describe('setAllNotificationRules', () => {
    it('sets list and status properties of state.', () => {
      const initialState = defaultNotificationRulesState

      const actual = rulesReducer(
        initialState,
        setAllNotificationRules(RemoteDataState.Done, [NEW_RULE_DRAFT])
      )

      const expected = {
        ...defaultNotificationRulesState,
        list: [NEW_RULE_DRAFT],
        status: RemoteDataState.Done,
      }

      expect(actual).toEqual(expected)
    })
  })

  describe('setRule', () => {
    it('adds rule to list if it is new', () => {
      const initialState = defaultNotificationRulesState

      const actual = rulesReducer(initialState, setRule(NEW_RULE_DRAFT))

      const expected = {
        ...defaultNotificationRulesState,
        list: [NEW_RULE_DRAFT],
      }

      expect(actual).toEqual(expected)
    })

    it('updates rule in list if it exists', () => {
      let initialState = defaultNotificationRulesState
      initialState.list = [NEW_RULE_DRAFT]

      const actual = rulesReducer(
        initialState,
        setRule({
          ...NEW_RULE_DRAFT,
          name: 'moo',
        })
      )

      const expected = {
        ...defaultNotificationRulesState,
        list: [{...NEW_RULE_DRAFT, name: 'moo'}],
      }

      expect(actual).toEqual(expected)
    })
  })

  describe('removeRule', () => {
    it('removes rule from list', () => {
      const initialState = defaultNotificationRulesState
      initialState.list = [NEW_RULE_DRAFT]
      const actual = rulesReducer(initialState, removeRule(NEW_RULE_DRAFT.id))

      const expected = {
        ...defaultNotificationRulesState,
        list: [],
      }

      expect(actual).toEqual(expected)
    })
  })

  describe('setCurrentRule', () => {
    it('sets current rule and status.', () => {
      const initialState = defaultNotificationRulesState

      const actual = rulesReducer(
        initialState,
        setCurrentRule(RemoteDataState.Done, NEW_RULE_DRAFT)
      )

      const expected = {
        ...defaultNotificationRulesState,
        current: {
          status: RemoteDataState.Done,
          rule: NEW_RULE_DRAFT,
        },
      }

      expect(actual).toEqual(expected)
    })
  })
})
