// Generated by https://pagedraw.io/pages/6103
import Folderlist from './folderlist';
import React from 'react';
import './folderlistelement.css';

function render() {
    return <div className="folderlistelement">
        { (this.props.state == "expanded") ?
            <div className="folderlistelement-0-0">
                <div className="folderlistelement-0-0-0">
                    <div className="folderlistelement-expanded-3">
                        <div className="folderlistelement-0-0-0-0-0">
                            <div className="folderlistelement-nestedlistelement-7">
                                <Folderlist name={this.props.name} contents={this.props.contents} /> 
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        : null}
        { (this.props.state == "collapsed") ?
            <div className="folderlistelement-1-0">
                <div className="folderlistelement-1-0-0">
                    <div className="folderlistelement-collapsed-4">
                        <div className="folderlistelement-1-0-0-0-0">
                            <div className="folderlistelement-name-1">
                                { this.props.name }
                            </div>
                            <div className="folderlistelement-plus-4">{"[+]"}</div>
                        </div>
                    </div>
                </div>
            </div>
        : null}
        { (this.props.state == "file") ?
            <div className="folderlistelement-2-0">
                <div className="folderlistelement-2-0-0">
                    <div className="folderlistelement-file-3">
                        <div className="folderlistelement-2-0-0-0-0">
                            <div className="folderlistelement-name-5">
                                { this.props.name }
                            </div>
                            <div className="folderlistelement-minus-4">{"[open]"}</div>
                        </div>
                    </div>
                </div>
            </div>
        : null}
        { (this.props.state == "expandedEmpty") ?
            <div className="folderlistelement-3-0">
                <div className="folderlistelement-3-0-0">
                    <div className="folderlistelement-expandedempty-0">
                        <div className="folderlistelement-3-0-0-0-0">
                            <div className="folderlistelement-name-13">
                                { this.props.name }
                            </div>
                            <div className="folderlistelement-minus-42">{"[-]"}</div>
                        </div>
                    </div>
                </div>
            </div>
        : null}
    </div>;
};

export default function(props) {
    return render.apply({props: props});
}