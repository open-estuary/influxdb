// Libraries
import React, {PureComponent, CSSProperties} from 'react'

// Components
import DatePicker from 'src/shared/components/dateRangePicker/DatePicker'
import {ClickOutside} from 'src/shared/components/ClickOutside'

// Types
import {TimeRange} from 'src/types'
import {Button, ComponentColor, ComponentSize} from '@influxdata/clockface'

interface Props {
  timeRange: TimeRange
  onSetTimeRange: (timeRange: TimeRange) => void
  onClose: () => void
  position?: {top?: number; right?: number; bottom?: number; left?: number}
}

interface State {
  lower: string
  upper: string
}

class DateRangePicker extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    const {
      timeRange: {lower, upper},
    } = props

    this.state = {lower, upper}
  }

  public render() {
    const {onClose} = this.props
    const {upper, lower} = this.state

    return (
      <ClickOutside onClickOutside={onClose}>
        <div
          className="range-picker react-datepicker-ignore-onclickoutside"
          style={this.stylePosition}
        >
          <button className="range-picker--dismiss" onClick={onClose} />
          <div className="range-picker--date-pickers">
            <DatePicker
              dateTime={lower}
              onSelectDate={this.handleSelectLower}
              label="Start"
            />
            <DatePicker
              dateTime={upper}
              onSelectDate={this.handleSelectUpper}
              label="Stop"
            />
          </div>
          <Button
            className="range-picker--submit"
            color={ComponentColor.Primary}
            size={ComponentSize.Small}
            onClick={this.handleSetTimeRange}
            text="Apply Time Range"
          />
        </div>
      </ClickOutside>
    )
  }

  private get stylePosition(): CSSProperties {
    const {position} = this.props

    if (!position) {
      return {
        top: `${window.innerHeight / 2}px`,
        left: `${window.innerWidth / 2}px`,
        transform: `translate(-50%, -50%)`,
      }
    }

    const style = Object.entries(position).reduce(
      (acc, [k, v]) => ({...acc, [k]: `${v}px`}),
      {}
    )

    return style
  }

  private handleSetTimeRange = (): void => {
    const {onSetTimeRange, timeRange} = this.props
    const {upper, lower} = this.state

    onSetTimeRange({...timeRange, lower, upper})
  }

  private handleSelectLower = (lower: string): void => {
    this.setState({lower})
  }

  private handleSelectUpper = (upper: string): void => {
    this.setState({upper})
  }
}

export default DateRangePicker
