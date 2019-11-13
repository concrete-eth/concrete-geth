// @flow

// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

import React, {Component} from 'react';

import withStyles from '@material-ui/core/styles/withStyles';
import Typography from '@material-ui/core/Typography';
import Grid from '@material-ui/core/Grid';
import ResponsiveContainer from 'recharts/es6/component/ResponsiveContainer';
import AreaChart from 'recharts/es6/chart/AreaChart';
import Area from 'recharts/es6/cartesian/Area';
import ReferenceLine from 'recharts/es6/cartesian/ReferenceLine';
import Label from 'recharts/es6/component/Label';
import Tooltip from 'recharts/es6/component/Tooltip';

import ChartRow from 'ChartRow';
import CustomTooltip, {bytePlotter, bytePerSecPlotter, percentPlotter, multiplier} from 'CustomTooltip';
import {chartStrokeWidth, styles as commonStyles} from '../common';
import type {General, System} from '../types/content';
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faNetworkWired} from "@fortawesome/free-solid-svg-icons";
import Toolbar from "@material-ui/core/Toolbar";

const FOOTER_SYNC_ID = 'footerSyncId';

const CPU     = 'cpu';
const MEMORY  = 'memory';
const DISK    = 'disk';
const TRAFFIC = 'traffic';

const TOP = 'Top';
const BOTTOM = 'Bottom';

const cpuLabelTop = 'Process load';
const cpuLabelBottom = 'System load';
const memoryLabelTop = 'Active memory';
const memoryLabelBottom = 'Virtual memory';
const diskLabelTop = 'Disk read';
const diskLabelBottom = 'Disk write';
const trafficLabelTop = 'Download';
const trafficLabelBottom = 'Upload';

// styles contains the constant styles of the component.
const styles = {
	footer: {
		maxWidth: '100%',
		flexWrap: 'nowrap',
		margin:   0,
	},
	chartRowWrapper: {
		height:  '100%',
		padding: 0,
	},
	doubleChartWrapper: {
		height: '100%',
		width:  '99%',
	},
	link: {
		color:          'inherit',
		textDecoration: 'none',
	},
};

// themeStyles returns the styles generated from the theme for the component.
const themeStyles: Object = (theme: Object) => ({
	footer: {
		backgroundColor: theme.palette.grey[900],
		color:           theme.palette.getContrastText(theme.palette.grey[900]),
		zIndex:          theme.zIndex.appBar,
		height:          theme.spacing.unit * 10,
	},
});

export type Props = {
	classes: Object, // injected by withStyles()
	theme: Object,
	general: General,
	system: System,
	shouldUpdate: Object,
};

type State = {};

// Footer renders the footer of the dashboard.
class Footer extends Component<Props, State> {
	shouldComponentUpdate(nextProps: Readonly<Props>, nextState: Readonly<State>, nextContext: any) {
		return typeof nextProps.shouldUpdate.general !== 'undefined' || typeof nextProps.shouldUpdate.system !== 'undefined';
	}

	// halfHeightChart renders an area chart with half of the height of its parent.
	halfHeightChart = (chartProps, tooltip, areaProps, label, position) => (
		<ResponsiveContainer width='100%' height='50%'>
			<AreaChart {...chartProps}>
				{!tooltip || (<Tooltip cursor={false} content={<CustomTooltip tooltip={tooltip} />} />)}
				<Area isAnimationActive={false} strokeWidth={chartStrokeWidth} type='monotone' {...areaProps} />
				<ReferenceLine x={0} strokeWidth={0}>
					<Label fill={areaProps.fill} value={label} position={position} />
				</ReferenceLine>
			</AreaChart>
		</ResponsiveContainer>
	);

	// doubleChart renders a pair of charts separated by the baseline.
	doubleChart = (syncId, chartKey, topChart, bottomChart) => {
		if (!Array.isArray(topChart.data) || !Array.isArray(bottomChart.data)) {
			return null;
		}
		const topDefault = topChart.default || 0;
		const bottomDefault = bottomChart.default || 0;
		const topKey = `${chartKey}${TOP}`;
		const bottomKey = `${chartKey}${BOTTOM}`;
		const topColor = '#8884d8';
		const bottomColor = '#82ca9d';

		return (
			<div style={styles.doubleChartWrapper}>
				{this.halfHeightChart(
					{
						syncId,
						data:   topChart.data.map(({value}) => ({[topKey]: value || topDefault})),
						margin: {top: 5, right: 5, bottom: 0, left: 5},
					},
					topChart.tooltip,
					{dataKey: topKey, stroke: topColor, fill: topColor},
					topChart.label,
					'insideBottomLeft',
				)}
				{this.halfHeightChart(
					{
						syncId,
						data:   bottomChart.data.map(({value}) => ({[bottomKey]: -value || -bottomDefault})),
						margin: {top: 0, right: 5, bottom: 5, left: 5},
					},
					bottomChart.tooltip,
					{dataKey: bottomKey, stroke: bottomColor, fill: bottomColor},
					bottomChart.label,
					'insideTopLeft',
				)}
			</div>
		);
	};

	render() {
		const {general, system} = this.props;
		let network = '';
		switch (general.genesis) {
		case '0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3':
			network = 'main';
			break;
		case '0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d':
			network = 'ropsten';
			break;
		case '0x6341fd3daf94b748c72ced5a5b26028f2474f5f00d824504e4fa37a75767e177':
			network = 'rinkeby';
			break;
		case '0xbf7e331f7f7c1dd2e05159666b3bf8bc7a8a3a9eb1d518969eab529dd9b88c1a':
			network = 'görli';
			break;
		default:
			network = `unknown (${general.genesis.substring(0, 8)})`;
		}

		return (
			<Grid container className={this.props.classes.footer} direction='row' alignItems='center' style={styles.footer}>
				<Grid item xs style={styles.chartRowWrapper}>
					<ChartRow>
						{this.doubleChart(
							FOOTER_SYNC_ID,
							CPU,
							{data: system.processCPU, tooltip: percentPlotter(cpuLabelTop), label: cpuLabelTop},
							{data: system.systemCPU, tooltip: percentPlotter(cpuLabelBottom, multiplier(-1)), label: cpuLabelBottom},
						)}
						{this.doubleChart(
							FOOTER_SYNC_ID,
							MEMORY,
							{data: system.activeMemory, tooltip: bytePlotter(memoryLabelTop), label: memoryLabelTop},
							{data: system.virtualMemory, tooltip: bytePlotter(memoryLabelBottom, multiplier(-1)), label: memoryLabelBottom},
						)}
						{this.doubleChart(
							FOOTER_SYNC_ID,
							DISK,
							{data: system.diskRead, tooltip: bytePerSecPlotter(diskLabelTop), label: diskLabelTop},
							{data: system.diskWrite, tooltip: bytePerSecPlotter(diskLabelBottom, multiplier(-1)), label: diskLabelBottom},
						)}
						{this.doubleChart(
							FOOTER_SYNC_ID,
							TRAFFIC,
							{data: system.networkIngress, tooltip: bytePerSecPlotter(trafficLabelTop), label: trafficLabelTop},
							{data: system.networkEgress, tooltip: bytePerSecPlotter(trafficLabelBottom, multiplier(-1)), label: trafficLabelBottom},
						)}
					</ChartRow>
				</Grid>
				<Grid item>
					<Typography type='caption' color='inherit'>
						<span style={commonStyles.light}>Geth</span> {general.version}
					</Typography>
					{general.commit && (
						<Typography type='caption' color='inherit'>
							<span style={commonStyles.light}>{'Commit '}</span>
							<a
								href={`https://github.com/ethereum/go-ethereum/commit/${general.commit}`}
								target='_blank'
								rel='noopener noreferrer'
								style={styles.link}
							>
								{general.commit.substring(0, 8)}
							</a>
						</Typography>
					)}
					<Typography style={styles.headerText}>
						<span style={commonStyles.light}>Network</span> {network}
					</Typography>
				</Grid>
			</Grid>
		);
	}
}

export default withStyles(themeStyles)(Footer);
