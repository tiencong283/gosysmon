import React from "react"
import "./IOCList.css"
import $ from "jquery"
import {Link} from "react-router-dom";
import PaginationNav from "../PaginationNav/PaginationNav";

const title = "IOC List - GoSysmon"
const endpoint = "/api/ioc"
const iocTypes = ["Hash", "IP", "Domain"]

class IOCList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            viewIOCs: [],
            iocList: [],
            paging: {   // pagination
                currentPageIdx: 0,
                elementsPerPage: 20,
                numOfPages: 0,
            }
        }
        this.handlePrevious = this.handlePrevious.bind(this)
        this.handleNext = this.handleNext.bind(this)
    }

    // pagination
    getViewElements(pageIdx) {
        return this.getViewElementsFrom(pageIdx, this.state.iocList)
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
            viewIOCs: this.getViewElements(newPageIdx),
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
            viewIOCs: this.getViewElements(newPageIdx),
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
                    viewIOCs: this.getViewElementsFrom(this.state.paging.currentPageIdx, data),
                    iocList: data,
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
                        <th>Timestamp</th>
                        <th>Type</th>
                        <th>Indicator</th>
                        <th>Notes</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.viewIOCs.map(function (ioc) {
                            return (
                                <tr>
                                    <td>{ioc.Timestamp}</td>
                                    <td>{iocTypes[ioc.IOCType]}</td>
                                    <td><a href={ioc.ExternalUrl}>{ioc.Indicator}</a></td>
                                    <td><Link
                                        to={`/process?ProviderGuid=${ioc.ProviderGuid}&ProcessGuid=${ioc.ProcessGuid}`}>
                                        {ioc.Message}</Link></td>
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

export default IOCList