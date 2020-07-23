import React from "react"
import "./HostList.css"
import $ from "jquery"
import PaginationNav from "../PaginationNav/PaginationNav";

const title = "Client List - GoSysmon"
const endpoint = "/api/host"

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
                        numOfPages: Math.floor(data.length / this.state.paging.elementsPerPage)
                    }
                })
            }.bind(this),
        })
    }

    render() {
        return (
            <div className="list-table-container">
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>
                <table className="list-table hover unstriped">
                    <thead>
                    <tr>
                        <th>Host Name</th>
                        <th>Status</th>
                        <th>Joined At</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.viewHosts.map(function (host) {
                            return (
                                <tr>
                                    <td><span>{host.Name}</span></td>
                                    <td><span>{host.Active ? "Active" : "Not active"}</span></td>
                                    <td><span>{host.FirstSeen}</span></td>
                                </tr>
                            )
                        })
                    }
                    </tbody>
                </table>
            </div>
        )
    }
}

export default HostList