import React from "react"
import "./ActivityLogList.css"

const title = "Activity Logs - GoSysmon"

class ActivityLogList extends React.Component {
    componentDidMount() {
        document.title = title
    }

    render() {
        return (
            <h1>ActivityLogList</h1>
        )
    }
}

export default ActivityLogList