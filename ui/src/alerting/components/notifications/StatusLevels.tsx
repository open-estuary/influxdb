// Libraries
import React, {FC} from 'react'

// Components
import {
  ComponentSpacer,
  TextBlock,
  FlexDirection,
  ComponentSize,
} from '@influxdata/clockface'
import LevelsDropdown from 'src/alerting/components/notifications/LevelsDropdown'
import StatusChangeDropdown from 'src/alerting/components/notifications/StatusChangeDropdown'
import {LevelType} from 'src/alerting/components/notifications/RuleOverlay.reducer'

// Utils
import {useRuleDispatch} from './RuleOverlay.reducer'

// Types
import {StatusRuleDraft, CheckStatusLevel} from 'src/types'

interface Props {
  status: StatusRuleDraft
}

const StatusLevels: FC<Props> = ({status}) => {
  const {currentLevel, previousLevel} = status.value
  const dispatch = useRuleDispatch()

  const onClickLevel = (levelType: LevelType, level: CheckStatusLevel) => {
    dispatch({
      type: 'UPDATE_STATUS_LEVEL',
      statusID: status.cid,
      levelType,
      level,
    })
  }

  return (
    <ComponentSpacer direction={FlexDirection.Row} margin={ComponentSize.Small}>
      <TextBlock text="When status" />
      <ComponentSpacer.FlexChild grow={0} basis={140}>
        <StatusChangeDropdown status={status} />
      </ComponentSpacer.FlexChild>
      {!!previousLevel && (
        <ComponentSpacer.FlexChild grow={0} basis={140}>
          <LevelsDropdown
            type="previousLevel"
            selectedLevel={previousLevel.level}
            onClickLevel={onClickLevel}
          />
        </ComponentSpacer.FlexChild>
      )}
      {!!previousLevel && <TextBlock text="to" />}
      <ComponentSpacer.FlexChild grow={0} basis={140}>
        <LevelsDropdown
          type="currentLevel"
          selectedLevel={currentLevel.level}
          onClickLevel={onClickLevel}
        />
      </ComponentSpacer.FlexChild>
    </ComponentSpacer>
  )
}

export default StatusLevels
