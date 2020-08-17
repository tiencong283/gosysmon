import React from "react"
import $ from "jquery"
import {Link} from "react-router-dom";
import PaginationNav from "../PaginationNav/PaginationNav";

const title = "IOC List - GoSysmon"
const endpoint = "/api/ioc"

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
                        numOfPages: Math.floor(data.length / this.state.paging.elementsPerPage) + 1
                    }
                })
            }.bind(this),
        })
    }

    render() {
        return (
            <div className="inner-content-wrapper">
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>

                <table className="common-table">
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
                        this.state.viewIOCs.map(function (ioc, idx) {
                            return (
                                <tr key={idx}>
                                    <td className="col-timestamp">{ioc.Timestamp}</td>
                                    <td>{ioc.IOCType}</td>
                                    <td><a href={ioc.ExternalUrl}>{ioc.Indicator}</a></td>
                                    <td className="col-notes"><Link to={ioc.ProcRefUrl}>{ioc.Message}</Link></td>
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