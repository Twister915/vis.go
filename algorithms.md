# Algorithms

To compute a single frame (at frame-rate *F*) of size *S* (in time units), the following procedure is used:
* *N*, the number of samples required, is computed. `N = sample_rate * (S / 1 second)`
* *N* samples are read, producing *N* data-points for each channel
* The file pointer is moved `((-S + (1 second / F)) / (1 second)) * sample_rate` samples back
* The value of a Blackman-Nuttall window of size *N* is applied to the data independently for each channel
* *C* (being the number of channels) Discrete Fourier Transforms are approximated using a Fast Fourier Transform algorithm producing *C*
  results *R[i]* (where i is channel number, and R[i] is the output of size `(N / 2) + 1`) in complex space
* The data is copied from all *R* to final *ℝ* in real space by calculating the *abs* of each complex value
* The output is "binned" into B bins, using the index computed by `B_i = floor(((f_i / f_max) ** (1 / gamma)) * B_max)`.
  Values belonging to the same bin are summed.
* All channels are averaged into a single value for the computation of the new µ and σ (old data is not destroyed, however)
* The values of µ and σ are updated to reflect the computed bin values, if needed (more detail on this in a different section)
* Using the µ and σ for the distribution of possible binned values, the extrema (anything below the
  32nd percentile and anything above the 97th percentile) are normalized linearly between [0.01, 1.00]
* Savitsky-Golay smoothing is applied to the values (adjacent bins, in this single frame will have their values influence one-another)
* In the case that this is not the first frame, exponential smoothing will be applied on a per-bin basis through time.
* The value for all current bins is copied to a buffer for the values from the "previous frame" (so that exponential
  time smoothing can be performed for the next frame)

In simple terms:
* The data is loaded from the wav file (moving the pointer forward by the size of each frame's window)
* A Blackman-Nuttall window is applied to each channel independently
* The fourier transform is computed for all data (for all channels)
* The complex result of the fourier transform (of size `(N / 2) + 1`) is transformed to the real space by taking the magnitude of each
  complex number.
* Data is binned according to `B_i = floor(((f_i / f_max) ** (1 / gamma)) * B_max)` where data is summed as it accumulates in each bin.
* The binned data (which is currently stored separately for each channel) is temporarily combined, so that the rolling average
  and std-deviation can be updated with the new data
* The data is normalized so that anything below the 32nd percentile will appear as 0.01 and anything above the 97th percentile will be 1.00
* Smoothing, within this frame, is performed so that the bins don't look too dissimilar to nearby bins
* Exponential-time-smoothing is performed, so that data cannot change wildly between frames.

When the program starts, to estimate the mean and std-deviation of the bar heights:
* Segments of audio of a fixed length (say: 800ms) are read and some frames are computed
  (if 800ms at 30fps then 37 frames will be computed). A "stride" between each segment is defined (say 5s), and this is the
  amount of time skipped after the end of each set of frames before computing the next one.
* A value µ and σ is computed for this "sample of the data."

Once that computation is complete, then the pointer in the audio file is reset to the beginning of the audio data, and
the computation for each frame will begin.