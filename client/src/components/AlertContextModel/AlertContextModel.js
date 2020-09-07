import React from "react"
import "./AlertContextModel.css"


export default class AlertContextModel extends React.Component {
    renderHeader() {
        if (!this.props.alert.Technique) {
            return <span></span>
        }
        return (
            <div id="alert-context-header">
                <a href={this.props.alert.Technique.Url}>Mitre ATT&CK <i className="fa fa-external-link"/></a>
            </div>
        )
    }

    renderPropList() {
        let alert = this.props.alert
        if (!alert.Context) {
            return <tr></tr>
        }
        let properties = Object.keys(alert.Context).map(function (key) {
            return [key, alert.Context[key]]
        })
        properties.sort(function (a, b) {   // order by property name
            if (a[0] > b[0]) {
                return 1
            }
            if (a[0] < b[0]) {
                return -1
            }
            return 0
        })
        return properties.map(function (arr, idx) {
            return (
                <tr key={idx}>
                    <td>{arr[0]}</td>
                    <td>{arr[1]}</td>
                </tr>
            )
        })
    }

    render() {
        if (!this.props.alert) {
            return <span></span>
        }
        return (
            <div id="alert-context">
                {
                    this.renderHeader()
                }
                <div id="alert-context-content">
                    <table className="common-table">
                        <thead>
                        <tr>
                            <th width="100">Property</th>
                            <th>Value</th>
                        </tr>
                        </thead>
                        <tbody>
                        {
                            this.renderPropList()
                        }
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }
}
