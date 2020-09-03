import React from "react"
import $ from "jquery";
import PaginationNav from "../PaginationNav/PaginationNav";
import Header from '../Header/Header'
import * as AuthService from "../Auth/AuthService";
import { Redirect } from "react-router-dom";

const title = "Activity Logs - GoSysmon"
const endpoint = "/api/activity-log"

class ActivityLog extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            viewActLogs: [],
            actLogs: [],
            paging: {   // pagination
                currentPageIdx: 0,
                elementsPerPage: 10,
                numOfPages: 0,
            }
        }
        this.handlePrevious = this.handlePrevious.bind(this)
        this.handleNext = this.handleNext.bind(this)
    }

    // pagination
    getViewElements(pageIdx) {
        return this.getViewElementsFrom(pageIdx, this.state.actLogs)
    }

    getViewElementsFrom(pageIdx, arr) {
        let startIdx = pageIdx * this.state.paging.elementsPerPage
        let endIdx = (pageIdx + 1) * this.state.paging.elementsPerPage
        return arr.slice(startIdx, endIdx)
    }

    handlePrevious(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx - 1
        if (newPageIdx < 0) {
            newPageIdx = 0
        }
        this.setState({
            viewActLogs: this.getViewElements(newPageIdx),
            paging: {
                ...this.state.paging,
                currentPageIdx: newPageIdx
            }
        })
    }

    handleNext(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx + 1
        if (newPageIdx >= this.state.paging.numOfPages) {
            newPageIdx = this.state.paging.numOfPages - 1
        }
        this.setState({
            viewActLogs: this.getViewElements(newPageIdx),
            paging: {
                ...this.state.paging,
                currentPageIdx: newPageIdx
            }
        })
    }

    componentDidMount() {
        document.title = title
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    viewActLogs: this.getViewElementsFrom(this.state.paging.currentPageIdx, data),
                    actLogs: data,
                    paging: {
                        ...this.state.paging,
                        numOfPages: Math.floor(data.length / this.state.paging.elementsPerPage) + 1
                    }
                })
            }.bind(this),
        })
    }

    render() {
        return (
            <div className="grid-container full">
                <div className="grid-x grid-margin-x main-container">
                    <Header />
                    <div className="cell auto content-wrapper">
                        <div className="inner-content-wrapper">
                            <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                                handleNext={this.handleNext} />
                            <table className="common-table">
                                <thead>
                                    <tr>
                                        <th>Timestamp</th>
                                        <th>Type</th>
                                        <th>Message</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {
                                        this.state.viewActLogs.map(function (actLog, idx) {
                                            return (
                                                <tr key={idx}>
                                                    <td className="col-timestamp"><span>{actLog.Timestamp}</span></td>
                                                    <td><span>{actLog.Type}</span></td>
                                                    <td className="col-notes"><span>{actLog.Message}</span></td>
                                                </tr>
                                            )
                                        })
                                    }
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

export default ActivityLog
