<?php

/**
 * This file is part of RoadRunner package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

declare(strict_types=1);

namespace Spiral\RoadRunner\Metrics;

use Spiral\RoadRunner\Metrics\Exception\MetricsException;

interface MetricsInterface
{
    /**
     * Add collector value. Fallback to appropriate method of related collector.
     *
     * @param string $name
     * @param float $value
     * @param array $labels
     *
     * @throws MetricsException
     */
    public function add(string $name, float $value, array $labels = []): void;

    /**
     * Subtract the collector value, only for gauge collector.
     *
     * @param string $name
     * @param float $value
     * @param array $labels
     *
     * @throws MetricsException
     */
    public function sub(string $name, float $value, array $labels = []): void;

    /**
     * Observe collector value, only for histogram and summary collectors.
     *
     * @param string $name
     * @param float $value
     * @param array $labels
     *
     * @throws MetricsException
     */
    public function observe(string $name, float $value, array $labels = []): void;

    /**
     * Set collector value, only for gauge collector.
     *
     * @param string $name
     * @param float $value
     * @param array $labels
     *
     * @throws MetricsException
     */
    public function set(string $name, float $value, array $labels = []): void;

    /**
     * Declares named collector.
     *
     * @param string $name
     * @param CollectorInterface $collector
     *
     * @throws MetricsException
     */
    public function declare(string $name, CollectorInterface $collector): void;
}
