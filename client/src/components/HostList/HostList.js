import React from "react"
import $ from "jquery"
import PaginationNav from "../PaginationNav/PaginationNav"
import Header from '../Header/Header'
import * as AuthService from "../Auth/AuthService";
import { Redirect } from "react-router-dom";


const title = "Client List - GoSysmon"
const endpoint = "/api/host"

const activeStatusColor = "rgb(49, 162, 76)"
const inactiveStatusColor = "rgb(230, 230, 250)"

class HostList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            viewHosts: [],
            hostList: [],
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
        return this.getViewElementsFrom(pageIdx, this.state.hostList)
    }

    getViewElementsFrom(pageIdx, iocList) {
        let startIdx = pageIdx * this.state.paging.elementsPerPage
        let endIdx = (pageIdx + 1) * this.state.paging.elementsPerPage
        return iocList.slice(startIdx, endIdx)
    }

    handlePrevious(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx - 1
        if (newPageIdx < 0) {
            newPageIdx = 0
        }
        this.setState({
            viewHosts: this.getViewElements(newPageIdx),
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
            viewHosts: this.getViewElements(newPageIdx),
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
                    viewHosts: this.getViewElementsFrom(this.state.paging.currentPageIdx, data),
                    hostList: data,
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
                                        <th>Host Name</th>
                                        <th>Status</th>
                                        <th>Joined At</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {
                                        this.state.viewHosts.map(function (host, idx) {
                                            return (
                                                <tr key={idx}>
                                                    <td><span>{host.Name}</span></td>
                                                    <td>
                                                        { /*https://iconscout.com/icon/check-1779337*/}
                                                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none"
                                                            viewBox="0 0 24 24">
                                                            <path fill={host.Active ? activeStatusColor : inactiveStatusColor}
                                                                fillRule="evenodd"
                                                                d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22ZM15.5355 8.46447C15.9261 8.07394 16.5592 8.07394 16.9498 8.46447C17.3403 8.85499 17.3403 9.48816 16.9498 9.87868L11.2966 15.5318L11.2929 15.5355C11.1919 15.6365 11.0747 15.7114 10.9496 15.7602C10.7724 15.8292 10.5795 15.8459 10.3948 15.8101C10.2057 15.7735 10.0251 15.682 9.87868 15.5355L9.87489 15.5317L7.05028 12.7071C6.65975 12.3166 6.65975 11.6834 7.05028 11.2929C7.4408 10.9024 8.07397 10.9024 8.46449 11.2929L10.5858 13.4142L15.5355 8.46447Z"
                                                                clipRule="evenodd" />
                                                        </svg>
                                                    </td>
                                                    <td className="col-timestamp"><span>{host.FirstSeen}</span></td>
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

export default HostList
