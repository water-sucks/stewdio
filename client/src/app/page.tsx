'use client';
import React, { useState } from 'react';
// Assuming these components are correctly imported from their paths
import WaveformCanvas from '../components/WaveformDisplay'; // Assuming this is the original single waveform display
import WaveformDifference from '../components/DifferenceDisplay'; // Assuming this is the difference display component
import FileSelector, { AudioFileItem } from '../components/FileSelector'; // Assuming this is your file selector

const Page = () => {
	// State to hold the selected audio files
	const [selectedFiles, setSelectedFiles] = useState<AudioFileItem[]>([]);

	// Determine the files to pass to WaveformDifference
	// Only pass files if at least two are selected
	const file1 = selectedFiles.length > 0 ? selectedFiles[0].file : null;
	const file2 = selectedFiles.length > 1 ? selectedFiles[1].file : null;

	return (
		<div style={{ display: 'flex', height: '100vh', fontFamily: 'sans-serif' }}>
			{/* Left - File Selector */}
			<div style={{ width: '300px', borderRight: '1px solid #ccc', padding: '1rem', overflowY: 'auto' }}>
				{/* Assuming FileSelector takes an onFileSelect prop */}
				<FileSelector onFileSelect={setSelectedFiles} />
			</div>

			{/* Right - Waveform Displays */}
			<div style={{
				flex: 1, // Take remaining space
				display: 'flex',
				flexDirection: 'column', // Stack waveforms vertically
				padding: '1rem',
				overflowY: 'auto', // Allow scrolling if content overflows
				gap: '1rem', // Add space between elements
			}}>
				{/* Display individual waveforms */}
				{selectedFiles.map((item, index) => (
					<div key={item.id} style={{ padding: '0.5rem' }}>
						<WaveformCanvas
							audioFile={item.file}
							// Assign distinct colors, e.g., based on index
							color={index % 2 === 0 ? '#2B5360' : '#764C7A'}
						/>
					</div>
				))}

				{/* Display the difference waveform only if two files are selected */}
				{file1 && file2 && (
					<div style={{ padding: '0.5rem', marginTop: '1rem' }}>
						<WaveformDifference
							audioFile1={file1} // Pass the first selected file
							audioFile2={file2} // Pass the second selected file
							positiveColor="#00FF00" // Optional: Green for positive diff
							negativeColor="#FF0000" // Optional: Red for negative diff
						/>
					</div>
				)}
			</div>
		</div>
	);
};

export default Page;
