<?php

/**
 * This file is part of RoadRunner package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Spiral\RoadRunner\Metrics;

use JetBrains\PhpStorm\Pure;

/**
 * @psalm-type CollectorType = CollectorInterface::TYPE_*
 *
 * @psalm-type ArrayFormatType = array {
 *      type:       CollectorType,
 *      namespace:  string,
 *      subsystem:  string,
 *      help:       string,
 *      labels:     array<array-key, string>,
 *      buckets:    array<array-key, float>
 * }
 */
interface CollectorInterface
{
    /**
     * @var string
     */
    public const TYPE_HISTOGRAM = 'histogram';

    /**
     * @var string
     */
    public const TYPE_GAUGE = 'gauge';

    /**
     * @var string
     */
    public const TYPE_COUNTER = 'counter';

    /**
     * @param string $namespace
     * @return self
     */
    #[Pure]
    public function withNamespace(string $namespace): self;

    /**
     * @param string $subsystem
     * @return self
     */
    #[Pure]
    public function withSubsystem(string $subsystem): self;

    /**
     * @param string $help
     * @return self
     */
    #[Pure]
    public function withHelp(string $help): self;

    /**
     * @param string ...$label
     * @return self
     */
    #[Pure]
    public function withLabels(string ...$label): self;

    /**
     * @return ArrayFormatType
     */
    #[Pure]
    public function toArray(): array;
}
